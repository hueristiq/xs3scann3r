package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	hqlog "github.com/hueristiq/hqgoutils/log"
	"github.com/hueristiq/hqgoutils/log/formatter"
	"github.com/hueristiq/hqgoutils/log/levels"
	"github.com/hueristiq/xs3scann3r/internal/configuration"
	"github.com/hueristiq/xs3scann3r/pkg/s3format"
	"github.com/logrusorgru/aurora/v3"
	"github.com/spf13/pflag"
)

var (
	au aurora.Aurora

	concurrency int
	dump        string

	input      string
	monochrome bool
	verbosity  string
)

func init() {
	// parse flags
	pflag.IntVarP(&concurrency, "concurrency", "c", 10, "")
	pflag.StringVarP(&dump, "dump", "p", "", "")
	pflag.StringVarP(&input, "input", "i", "", "")
	pflag.BoolVarP(&monochrome, "monochrome", "m", false, "")
	pflag.StringVarP(&verbosity, "verbosity", "v", string(levels.LevelInfo), "")

	pflag.CommandLine.SortFlags = false
	pflag.Usage = func() {
		fmt.Fprintln(os.Stderr, configuration.BANNER)

		h := "USAGE:\n"
		h += "  xs3scann3r [OPTIONS]\n"

		h += "\nINPUT:\n"
		h += "  -i, --input         input file (use `-` to get from stdin)\n"

		h += "\nCONFIGURATIONS:\n"
		h += "   -c, --concurrency  number of concurrent threads (default: 10)\n"
		h += "   -d, --dump         location to dump objects\n"

		h += "\nOUTPUT:\n"
		h += "  -m, --monochrome    disable output content coloring\n"
		h += fmt.Sprintf("  -v, --verbosity     debug, info, warning, error, fatal or silent (default: %s)\n", string(levels.LevelInfo))

		fmt.Fprint(os.Stderr, h)
	}

	pflag.Parse()

	// initialize logger
	hqlog.DefaultLogger.SetMaxLevel(levels.LevelStr(verbosity))
	hqlog.DefaultLogger.SetFormatter(formatter.NewCLI(&formatter.CLIOptions{
		Colorize: !monochrome,
	}))

	au = aurora.NewAurora(!monochrome)
}

func main() { //nolint:gocyclo // To be refactored
	// input s3 buckets
	buckets := make(chan string, concurrency)

	go func() {
		defer close(buckets)

		var (
			err  error
			file *os.File
			stat fs.FileInfo
		)

		switch {
		case input != "" && input == "-":
			stat, err = os.Stdin.Stat()
			if err != nil {
				hqlog.Fatal().Msg("no stdin")
			}

			if stat.Mode()&os.ModeNamedPipe == 0 {
				hqlog.Fatal().Msg("no stdin")
			}

			file = os.Stdin
		case input != "" && input != "-":
			file, err = os.Open(input)
			if err != nil {
				hqlog.Fatal().Msg(err.Error())
			}
		default:
			hqlog.Fatal().Msg("xs3scann3r takes input from stdin or file using ")
		}

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			bucket := scanner.Text()

			if bucket != "" {
				buckets <- bucket
			}
		}

		if err := scanner.Err(); err != nil {
			hqlog.Fatal().Msg(err.Error())
		}
	}()

	wg := &sync.WaitGroup{}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			cfg, err := config.LoadDefaultConfig(context.TODO())
			if err != nil {
				hqlog.Fatal().Msg(err.Error())
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
				GetBucketACLInput := &s3.GetBucketAclInput{
					Bucket: aws.String(bucket),
				}

				GetBucketACLOutput, err := client.GetBucketAcl(context.TODO(), GetBucketACLInput)
				if err != nil {
					logger.Printf("GET ACL: %s\n", au.BrightRed("Failed").Bold())
				} else {
					GROUPS := map[string]string{
						"http://acs.amazonaws.com/groups/global/AllUsers":           "Everyone",
						"http://acs.amazonaws.com/groups/global/AuthenticatedUsers": "Authenticated AWS users",
					}
					PERMISSIONS := map[string][]string{}

					for _, grant := range GetBucketACLOutput.Grants {
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
					if dump != "" {
						// create the directory
						directory := filepath.Join(dump, bucket)

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
