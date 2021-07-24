package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/logrusorgru/aurora/v3"
	"github.com/signedsecurity/sigs3scann3r/pkg/s3format"
)

type options struct {
	inputList        string
	concurrency      int
	noColor, verbose bool
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
	flag.StringVar(&co.inputList, "input-list", "", "")
	flag.StringVar(&co.inputList, "iL", "", "")
	flag.IntVar(&co.concurrency, "concurrency", 10, "")
	flag.IntVar(&co.concurrency, "c", 10, "")
	flag.BoolVar(&co.noColor, "nC", false, "")
	flag.BoolVar(&co.verbose, "v", false, "")

	flag.Usage = func() {
		banner()

		h := "USAGE:\n"
		h += "  sigs3scann3r [OPTIONS]\n"

		h += "\nOPTIONS:\n"
		h += "  -iL, --input-list   input buckets list (use `iL -` to read from stdin)\n"
		h += "   -c, --concurrency  number of concurrent threads (default: 10)\n"
		h += "  -nC, --no-color     no color mode (default: false)\n"
		h += "   -v, --verbose      verbose mode\n"

		fmt.Fprint(os.Stderr, h)
	}

	flag.Parse()

	au = aurora.NewAurora(!co.noColor)
}

func main() {
	buckets := make(chan string, co.concurrency)

	go func() {
		defer close(buckets)

		var scanner *bufio.Scanner

		if co.inputList == "-" {
			stat, err := os.Stdin.Stat()
			if err != nil {
				log.Fatalln(errors.New("no stdin"))
			}

			if stat.Mode()&os.ModeNamedPipe == 0 {
				log.Fatalln(errors.New("no stdin"))
			}

			scanner = bufio.NewScanner(os.Stdin)
		} else {
			openedFile, err := os.Open(co.inputList)
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

	wg := &sync.WaitGroup{}

	for i := 0; i < co.concurrency; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			cfg, err := config.LoadDefaultConfig(context.TODO())
			if err != nil {
				log.Fatalln(err)
			}

			for bucket := range buckets {
				logger := log.New(os.Stdout, fmt.Sprintf(" %s | ", s3format.ToName(bucket)), 0)

				// Check existence & Get Region
				res, err := http.Get("http://" + s3format.ToVHost(bucket))
				if err != nil {
					fmt.Println(err)
					continue
				}

				if res.StatusCode == http.StatusNotFound {
					logger.Printf("STATUS: %s\n", au.BrightRed("Not Found").Bold())
					continue
				} else {
					logger.Printf("STATUS: %s\n", au.BrightGreen("Found").Bold())
				}

				// Extract Region
				region := res.Header.Get("X-Amz-Bucket-Region")

				logger.Printf("REGION: %s\n", region)

				// New Client
				client := s3.NewFromConfig(cfg, func(o *s3.Options) {
					o.Region = region
				})

				// GetBucketAcl

				GetBucketAclInput := &s3.GetBucketAclInput{
					Bucket: aws.String(s3format.ToName(bucket)),
				}

				GetBucketAclOutput, err := client.GetBucketAcl(context.TODO(), GetBucketAclInput)
				if err != nil {
					logger.Printf("GET ACL: %s\n", au.BrightRed("Failed").Bold())
				} else {
					GROUPS := map[string]string{
						"http://acs.amazonaws.com/groups/global/AllUsers":           "Everyone",
						"http://acs.amazonaws.com/groups/global/AuthenticatedUsers": "Authenticated AWS users",
					}
					PERMISSIONS := map[string][]string{}

					for _, grant := range GetBucketAclOutput.Grants {
						if grant.Grantee.Type == "Group" {
							for GROUP := range GROUPS {
								if *grant.Grantee.URI == GROUP {
									PERMISSIONS[GROUPS[GROUP]] = append(PERMISSIONS[GROUPS[GROUP]], string(grant.Permission))
								}
							}
						}
					}

					ACL := []string{}

					for PERMISSION := range PERMISSIONS {
						ACL = append(ACL, fmt.Sprintf("%s: %s", PERMISSION, strings.Join(PERMISSIONS[PERMISSION], ", ")))
					}

					logger.Printf("GET ACL: %s\n", strings.Join(ACL, "; "))
				}

				// PutObject

				PutObjectInput := &s3.PutObjectInput{
					Bucket: aws.String(s3format.ToName(bucket)),
					Key:    aws.String("etetst.txt"),
				}

				_, err = client.PutObject(context.TODO(), PutObjectInput)
				if err != nil {
					logger.Printf("PUT OBJECTS: %s\n", au.BrightRed("Failed").Bold())
				} else {
					logger.Printf("PUT OBJECTS: %s\n", au.BrightGreen("Success").Bold())
				}

				// ListObjectsV2

				ListObjectsV2Input := &s3.ListObjectsV2Input{
					Bucket: aws.String(s3format.ToName(bucket)),
				}

				_, err = client.ListObjectsV2(context.TODO(), ListObjectsV2Input)
				if err != nil {
					logger.Printf("GET OBJECTS: %s\n", au.BrightRed("Failed").Bold())
				} else {
					logger.Printf("GET OBJECTS: %s\n", au.BrightGreen("Success").Bold())
				}
				// if ListObjectsV2Output != nil && ListObjectsV2Output.Contents != nil {
				// 	fmt.Println("   ", au.BrightGreen("+").Bold(), "Objects:")

				// 	for _, item := range ListObjectsV2Output.Contents {
				// 		fmt.Println("       ", au.BrightGreen("+").Bold(), *item.Key, au.BrightGreen("size:"), item.Size, au.BrightGreen("last_modified:"), *item.LastModified)
				// 	}
				// }
			}
		}()
	}

	wg.Wait()
}
