package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	gofofa "github.com/FofaInfo/GoFOFA"
	"github.com/FofaInfo/GoFOFA/pkg/outformats"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	batchType string // batch query, can be ip/domain
)

const (
	IPMax     = 100 // maximum number of splicing ips
	DomainMax = 50  // maximum number of splicing domains
)

// dump subcommand
var dumpCmd = &cli.Command{
	Name:                   "dump",
	Usage:                  "fofa dump data",
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "fields",
			Aliases:     []string{"f"},
			Value:       "host,ip,port",
			Usage:       "visit fofa website for more info",
			Destination: &fieldString,
		},
		&cli.StringFlag{
			Name:        "format",
			Value:       "csv",
			Usage:       "can be csv/json/xml",
			Destination: &format,
		},
		&cli.BoolFlag{
			Name:        "json",
			Aliases:     []string{"j"},
			Usage:       "output use json format",
			Destination: &json,
		},
		&cli.StringFlag{
			Name:        "outFile",
			Aliases:     []string{"o"},
			Usage:       "if not set, write to stdout",
			Destination: &outFile,
		},
		&cli.StringFlag{
			Name:        "inFile",
			Aliases:     []string{"i"},
			Usage:       "queries line by line",
			Destination: &inFile,
		},
		&cli.IntFlag{
			Name:        "size",
			Aliases:     []string{"s"},
			Value:       -1,
			Usage:       "-1 means all",
			Destination: &size,
		},
		&cli.BoolFlag{
			Name:        "fixUrl",
			Value:       false,
			Usage:       "each host fix as url, like 1.1.1.1,80 will change to http://1.1.1.1",
			Destination: &fixUrl,
		},
		&cli.StringFlag{
			Name:        "urlPrefix",
			Value:       "http://",
			Usage:       "prefix of url, default is http://, can be redis:// and so on ",
			Destination: &urlPrefix,
		},
		&cli.BoolFlag{
			Name:        "full",
			Value:       false,
			Usage:       "search result for over a year",
			Destination: &full,
		},
		&cli.IntFlag{
			Name:        "batchCount",
			Aliases:     []string{"bc"},
			Value:       10,
			Usage:       "the count of ip/domain to query in each batch",
			Destination: &batchCount,
		},
		&cli.IntFlag{
			Name:        "batchSize",
			Aliases:     []string{"bs"},
			Value:       1000,
			Usage:       "the amount of data contained in each batch",
			Destination: &batchSize,
		},
		&cli.StringFlag{
			Name:        "batchType",
			Aliases:     []string{"bt"},
			Value:       "",
			Usage:       "batch query, can be ip/domain",
			Destination: &batchType,
		},
		&cli.BoolFlag{
			Name:        "headline",
			Value:       false,
			Usage:       "add headline for csv",
			Destination: &headline,
		},
		&cli.StringFlag{
			Name:        "customFields",
			Aliases:     []string{"cf"},
			Value:       "",
			Usage:       "use custom fields",
			Destination: &customFields,
		},
	},
	Action: DumpAction,
}

func constructQuery(queryType string, queries []string) string {
	var queryBuilder strings.Builder
	for i, query := range queries {
		if i > 0 {
			queryBuilder.WriteString(" || ")
		}
		queryBuilder.WriteString(fmt.Sprintf("%s==%s", queryType, query))
	}
	return queryBuilder.String()
}

func batchProcess(queries []string, batchSize int, queryType string) []string {
	var batchedQueries []string
	for i := 0; i < len(queries); i += batchSize {
		end := i + batchSize
		if end > len(queries) {
			end = len(queries)
		}
		batchedQuery := constructQuery(queryType, queries[i:end])
		batchedQueries = append(batchedQueries, batchedQuery)
	}
	return batchedQueries
}

// DumpAction search action
func DumpAction(ctx *cli.Context) error {
	// valid same config
	var queries []string
	query := ctx.Args().First()
	if len(query) > 0 {
		queries = append(queries, query)
	}
	if len(inFile) > 0 {
		// 打开文件
		file, err := os.Open(inFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		// 创建一个 Scanner 对象来逐行读取文件
		scanner := bufio.NewScanner(file)

		// 逐行读取并打印
		for scanner.Scan() {
			if scanner.Text() == "" {
				continue
			}
			queries = append(queries, scanner.Text())
		}

		// 检查是否有读取错误
		if err = scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}

	if len(queries) == 0 {
		return errors.New("fofa query cannot be empty, use args or -inFile")
	}

	var err error
	if customFields != "" {
		fieldString, err = getCustomFields(customFields)
		if err != nil {
			return fmt.Errorf("get custom fields error, %v", err)
		}
	}

	fields := strings.Split(fieldString, ",")
	if len(fields) == 0 {
		return errors.New("fofa fields cannot be empty")
	}

	// headline只允许在format=csv的情况下使用
	if headline && format != "csv" && len(outFile) > 0 {
		return errors.New("headline param is only allowed if format is csv, outFile not be empty")
	}

	// batchType检验
	if batchType != "" && batchType != "ip" && batchType != "domain" {
		return errors.New("batchType param has to be one of ip/domain")
	}

	if batchType == "ip" {
		if batchCount <= IPMax {
			queries = batchProcess(queries, batchCount, "ip")
		} else {
			queries = batchProcess(queries, IPMax, "ip")
		}
	} else if batchType == "domain" {
		if batchCount <= DomainMax {
			queries = batchProcess(queries, batchCount, "domain")
		} else {
			queries = batchProcess(queries, DomainMax, "domain")
		}
	}

	// gen output
	var outTo io.Writer
	if len(outFile) > 0 {
		var f *os.File
		if f, err = os.Create(outFile); err != nil {
			return fmt.Errorf("create outFile %s failed: %w", outFile, err)
		}
		outTo = f
		defer f.Close()
	} else {
		outTo = os.Stdout
	}

	if json {
		format = "json"
	}
	// gen writer
	var writer outformats.OutWriter
	if hasBodyField(fields) && format == "csv" {
		logrus.Warnln("fields contains body, so change format to json")
		writer = outformats.NewJSONWriter(outTo, fields)
	} else {
		switch format {
		case "csv":
			writer = outformats.NewCSVWriter(outTo)
		case "json":
			writer = outformats.NewJSONWriter(outTo, fields)
		case "xml":
			writer = outformats.NewXMLWriter(outTo, fields)
		default:
			return fmt.Errorf("unknown format: %s", format)
		}
	}

	if headline && format == "csv" && len(outFile) > 0 {
		// 写入表头
		err := writer.WriteAll([][]string{fields})
		if err != nil {
			return err
		}
	}

	// do search
	for _, query := range queries {
		log.Println("dump data of query:", query)

		fetchedSize := 0
		err := fofaCli.DumpSearch(query, size, batchSize, fields, func(res [][]string, allSize int) (err error) {
			fetchedSize += len(res)
			log.Printf("size: %d/%d, %.2f%%", fetchedSize, allSize, 100*float32(fetchedSize)/float32(allSize))
			// output
			err = writer.WriteAll(res)
			return err
		}, gofofa.SearchOptions{
			FixUrl:    fixUrl,
			UrlPrefix: urlPrefix,
			Full:      full,
		})
		if err != nil {
			log.Println("fetch error:", err)
			//return err
		}
	}

	return nil
}
