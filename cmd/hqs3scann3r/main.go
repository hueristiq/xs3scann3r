package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hueristiq/hqs3scann3r/pkg/s3format"
	"github.com/logrusorgru/aurora/v3"
)

type options struct {
	concurrency int
	dump        string
	inputList   string
	noColor     bool
	verbose     bool
}

var (
	co options
	au aurora.Aurora
)

func banner() {
	fmt.Fprintln(os.Stderr, aurora.BrightBlue(`
 _               _____                           _____      
| |__   __ _ ___|___ / ___  ___ __ _ _ __  _ __ |___ / _ __ 
| '_ \ / _`+"`"+` / __| |_ \/ __|/ __/ _`+"`"+` | '_ \| '_ \  |_ \| '__|
| | | | (_| \__ \___) \__ \ (_| (_| | | | | | | |___) | |   
|_| |_|\__, |___/____/|___/\___\__,_|_| |_|_| |_|____/|_| v1.1.0
          |_|
`).Bold())
}

func init() {
	flag.IntVar(&co.concurrency, "concurrency", 10, "")
	flag.IntVar(&co.concurrency, "c", 10, "")
	flag.StringVar(&co.dump, "dump", "", "")
	flag.StringVar(&co.dump, "d", "", "")
	flag.StringVar(&co.inputList, "input-list", "", "")
	flag.StringVar(&co.inputList, "iL", "", "")
	flag.BoolVar(&co.noColor, "no-color", false, "")
	flag.BoolVar(&co.noColor, "nC", false, "")
	flag.BoolVar(&co.verbose, "verbose", false, "")
	flag.BoolVar(&co.verbose, "v", false, "")

	flag.Usage = func() {
		banner()

		h := "USAGE:\n"
		h += "  hqs3scann3r [OPTIONS]\n"

		h += "\nOPTIONS:\n"
		h += "   -c, --concurrency  number of concurrent threads (default: 10)\n"
		h += "   -d, --dump         location to dump objects\n"
		h += "  -iL, --input-list   buckets list (use `-iL -` to read from stdin)\n"
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
				bucket = s3format.ToName(bucket)

				logger := log.New(os.Stdout, fmt.Sprintf(" %s | ", bucket), 0)

				// GetBucketRegion
				region, err := manager.GetBucketRegion(context.TODO(), s3.NewFromConfig(cfg), bucket)
				if err != nil {
					var bnf manager.BucketNotFound

					if errors.As(err, &bnf) {
						logger.Printf("STATUS: %s\n", au.BrightRed("Not Found").Bold())
						continue
					}

					log.Println("error:", err)
					continue
				}

				logger.Printf("STATUS: %s\n", au.BrightGreen("Found").Bold())
				logger.Printf("REGION: %s\n", region)

				// New Client
				client := s3.NewFromConfig(cfg, func(o *s3.Options) {
					o.Region = region
				})

				// GetBucketAcl
				GetBucketAclInput := &s3.GetBucketAclInput{
					Bucket: aws.String(bucket),
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
					Bucket: aws.String(bucket),
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
					Bucket: aws.String(bucket),
				}

				ListObjectsV2Output, err := client.ListObjectsV2(context.TODO(), ListObjectsV2Input)
				if err != nil {
					logger.Printf("GET OBJECTS: %s\n", au.BrightRed("Failed").Bold())
				} else {
					logger.Printf("GET OBJECTS: %s\n", au.BrightGreen("Success").Bold())
				}

				if ListObjectsV2Output != nil && ListObjectsV2Output.Contents != nil {
					if co.dump != "" {
						// create the directory
						directory := filepath.Join(co.dump, bucket)

						if _, err := os.Stat(directory); os.IsNotExist(err) {
							if directory != "" {
								err = os.MkdirAll(directory, os.ModePerm)
								if err != nil {
									fmt.Println(err)
									continue
								}
							}
						}

						// Dump objects
						downloader := manager.NewDownloader(client)

						for _, object := range ListObjectsV2Output.Contents {
							// Create the directories in the path
							file := filepath.Join(directory, aws.ToString(object.Key))

							if err := os.MkdirAll(filepath.Dir(file), 0775); err != nil {
								log.Println(err)
								continue
							}

							// Set up the local file
							fd, err := os.Create(file)
							if err != nil {
								log.Println(err)
								continue
							}

							// Download the file using the AWS SDK for Go
							_, err = downloader.Download(
								context.TODO(),
								fd,
								&s3.GetObjectInput{
									Bucket: aws.String(bucket),
									Key:    object.Key,
								},
							)
							if err != nil {
								log.Println(err)
								continue
							}

							fd.Close()
						}
					}
				}
			}
		}()
	}

	wg.Wait()
}
