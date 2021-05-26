package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/logrusorgru/aurora/v3"
	"github.com/signedsecurity/sigs3scann3r/pkg/sigs3scann3r"
)

type options struct {
	buckets string
	dump    bool
	noColor bool
	output  string
	verbose bool
}

var (
	co options
	au aurora.Aurora
)

func banner() {
	fmt.Fprintln(os.Stderr, aurora.BrightBlue(`
     _           _____                           _____
 ___(_) __ _ ___|___ / ___  ___ __ _ _ __  _ __ |___ / _ __
/ __| |/ _`+"`"+` / __| |_ \/ __|/ __/ _`+"`"+` | '_ \| '_ \  |_ \| '__|
\__ \ | (_| \__ \___) \__ \ (_| (_| | | | | | | |___) | |
|___/_|\__, |___/____/|___/\___\__,_|_| |_|_| |_|____/|_| v1.0.0
       |___/
`).Bold())
}

func init() {
	flag.BoolVar(&co.dump, "dump", false, "")
	flag.StringVar(&co.buckets, "iL", "", "")
	flag.BoolVar(&co.noColor, "nC", false, "")
	flag.StringVar(&co.output, "o", "./buckets", "")
	flag.BoolVar(&co.verbose, "v", false, "")

	flag.Usage = func() {
		banner()

		h := "USAGE:\n"
		h += "  sigs3scann3r [OPTIONS]\n"

		h += "\nOPTIONS:\n"
		h += "  -dump          dump found open buckets locally (default: false)\n"
		h += "  -iL            input buckets list (use `iL -` to read from stdin)\n"
		h += "  -nC            no color mode (default: false)\n"
		h += "  -o             buckets dump directory (default: ./buckets)\n"
		h += "  -v             verbose mode\n"

		fmt.Fprint(os.Stderr, h)
	}

	flag.Parse()

	au = aurora.NewAurora(!co.noColor)
}

func main() {
	buckets := make(chan string)

	go func() {
		defer close(buckets)

		var scanner *bufio.Scanner

		if co.buckets == "-" {
			stat, err := os.Stdin.Stat()
			if err != nil {
				log.Fatalln(errors.New("no stdin"))
			}

			if stat.Mode()&os.ModeNamedPipe == 0 {
				log.Fatalln(errors.New("no stdin"))
			}

			scanner = bufio.NewScanner(os.Stdin)
		} else {
			openedFile, err := os.Open(co.buckets)
			if err != nil {
				log.Fatalln(err)
			}
			defer openedFile.Close()

			scanner = bufio.NewScanner(openedFile)
		}

		for scanner.Scan() {
			if scanner.Text() != "" {
				buckets <- scanner.Text()
			}
		}

		if scanner.Err() != nil {
			log.Fatalln(scanner.Err())
		}
	}()

	for bucket := range buckets {
		// Check existence & Get Region
		res, err := http.Get("http://" + sigs3scann3r.Format(bucket, "vhost"))
		if err != nil {
			fmt.Println(err)
			continue
		}

		defer res.Body.Close()

		if res.StatusCode == http.StatusNotFound {
			fmt.Println(au.BrightRed("-").Bold(), sigs3scann3r.Format(bucket, "name"), "[", au.BrightRed("Not Found").Bold(), "]")
			continue
		}

		fmt.Println(au.BrightGreen("+").Bold(), sigs3scann3r.Format(bucket, "name"))

		// Extract Region
		region := res.Header.Get("X-Amz-Bucket-Region")

		fmt.Println("   ", au.BrightGreen("+").Bold(), "Region:", region)

		scanner, err := sigs3scann3r.New(region)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Get bucket ACL
		aclResult, err := scanner.Service.GetBucketAcl(&s3.GetBucketAclInput{
			Bucket: aws.String(sigs3scann3r.Format(bucket, "name")),
		})
		if err != nil {
			errorf(err.Error(), co.verbose)

			ERRORS := []string{"AccessDenied", "AllAccessDisabled"}

			for _, ERROR := range ERRORS {
				if strings.Contains(fmt.Sprintln(err), ERROR) {
					fmt.Println("   ", au.BrightRed("-").Bold(), "ACL:", ERROR)
					break
				}
			}
		} else {
			GROUPS := map[string]string{
				"http://acs.amazonaws.com/groups/global/AllUsers":           "Everyone",
				"http://acs.amazonaws.com/groups/global/AuthenticatedUsers": "Authenticated AWS users",
			}
			PERMISSIONS := map[string][]string{}

			for _, grant := range aclResult.Grants {
				if *grant.Grantee.Type == "Group" {
					for GROUP := range GROUPS {
						if *grant.Grantee.URI == GROUP {
							PERMISSIONS[GROUPS[GROUP]] = append(PERMISSIONS[GROUPS[GROUP]], *grant.Permission)
						}
					}
				}
			}

			fmt.Println("   ", au.BrightGreen("+").Bold(), "ACL:")
			for PERMISSION := range PERMISSIONS {
				fmt.Println("       ", au.BrightGreen("+").Bold(), PERMISSION, ":", strings.Join(PERMISSIONS[PERMISSION], ", "))
			}
		}

		// List Objects
		objectsResults, err := scanner.Service.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
		if err != nil {
			errorf(err.Error(), co.verbose)
		}

		if len(objectsResults.Contents) > 0 {
			fmt.Println("   ", au.BrightGreen("+").Bold(), "Objects:")

			for _, item := range objectsResults.Contents {
				fmt.Println("       ", au.BrightGreen("+").Bold(), *item.Key, au.BrightGreen("size:"), *item.Size, au.BrightGreen("last_modified:"), *item.LastModified)

				if co.dump {
					output := co.output + "/" + sigs3scann3r.Format(bucket, "name") + "/" + *item.Key

					if _, err := os.Stat(output); os.IsNotExist(err) {
						directory, _ := path.Split(output)

						if _, err := os.Stat(directory); os.IsNotExist(err) {
							if directory != "" {
								if err = os.MkdirAll(directory, os.ModePerm); err != nil {
									log.Fatalln(err)
								}
							}
						}
					}

					file, err := os.Create(output)
					if err != nil {
						errorf(err.Error(), co.verbose)
					}

					defer file.Close()

					numBytes, err := scanner.Downloader.Download(file,
						&s3.GetObjectInput{
							Bucket: aws.String(bucket),
							Key:    aws.String(*item.Key),
						})
					if err != nil {
						errorf(err.Error(), co.verbose)
					}

					fmt.Println("Downloaded", file.Name(), numBytes, "bytes")
				}
			}
		}
	}
}

func errorf(msg string, verbose bool, args ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stderr, msg+"\n", args...)
	}
}
