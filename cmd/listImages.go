package cmd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

type CatalogResponse struct {
	Repositories []string `json:"repositories"`
}

type TagsResponse struct {
	Tags         []string `json:"tags"`
	CreatedTime  []string `json:"createdTime"`  // 添加创建时间字段
	Architecture []string `json:"architecture"` //添加架构字段
}

type ManifestResponse struct {
	Config       Config `json:"config"`
	Architecture Config `json:"architecture"`
}

type Config struct {
	Digest string `json:"digest"`
}

type ConfigBlob struct {
	Created string `json:"created"`
}

type ManifestList struct {
	Manifests    []Manifest `json:"manifests"`
	Architecture string     `json:"architecture"`
}

type Manifest struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int    `json:"size"`
	Platform  Platform
}

type Platform struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
}

func listImagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all images and tags in Docker private registry",
		RunE:  listImages,
	}

	cmd.Flags().String("url", "", "Docker registry URL")
	cmd.Flags().BoolP("all", "a", false, "Show all tags")
	cmd.Flags().String("s", "", "Search for images using prefix matching.")

	return cmd
}

func listImages(cmd *cobra.Command, args []string) error {
	//获取table
	t := getTable()

	registryURL, _ := cmd.Flags().GetString("url")
	if registryURL == "" {
		return fmt.Errorf("Registry URL is required")
	}
	sName, _ := cmd.Flags().GetString("s")
	showAll, _ := cmd.Flags().GetBool("all")

	// 创建一个自定义的 http.Client，并禁用证书验证
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// 获取仓库的镜像名
	catalogURL := fmt.Sprintf("%s/v2/_catalog", registryURL)
	catalogResponse, err := httpClient.Get(catalogURL)
	if err != nil {
		return err
	}
	defer catalogResponse.Body.Close()

	body, err := ioutil.ReadAll(catalogResponse.Body)
	if err != nil {
		return err
	}

	var catalog CatalogResponse
	err = json.Unmarshal(body, &catalog)
	if err != nil {
		return err
	}

	// 遍历获取 tags
	for _, repo := range catalog.Repositories {
		if sName == "" || strings.HasPrefix(repo, sName) {
			t, _ = getTags(httpClient, registryURL, repo, showAll, t)
		} else {
		}
	}

	fmt.Println(t.Render())

	return nil
}

func getTable() table.Table {
	// 创建 table
	t := table.Table{}
	//设置表格标题
	header := table.Row{"image", "tags", "architecture", "Created Time"}
	t.AppendHeader(header)
	//加表格分割线
	//t.Style().Options.SeparateRows = true
	//设置居中
	column := []string{"image", "tags", "architecture", "Created Time"}
	c := []table.ColumnConfig{}
	// 根据表格的列数循环进行设置，统一居中
	for i := 0; i < len(column); i++ {
		name := column[i]
		if name == "image" {
			c = append(c, table.ColumnConfig{
				Name: "image",
				//AutoMerge:   true,
				VAlign:      text.VAlignMiddle, // 这里是垂直居中
				Align:       text.AlignCenter,
				AlignHeader: text.AlignCenter,
				AlignFooter: text.AlignCenter,
			})
			continue
		}
		c = append(c, table.ColumnConfig{
			Name:        column[i],
			Align:       text.AlignCenter,
			AlignHeader: text.AlignCenter,
			AlignFooter: text.AlignCenter,
		})
	}
	t.SetColumnConfigs(c)

	return t
}

func getTags(httpClient *http.Client, registryURL string, repo string, showAll bool, t table.Table) (table.Table, error) {
	tagsURL := fmt.Sprintf("%s/v2/%s/tags/list", registryURL, repo)
	tagsResponse, err := httpClient.Get(tagsURL)
	if err != nil {
		return t, err
	}
	defer tagsResponse.Body.Close()

	body, err := ioutil.ReadAll(tagsResponse.Body)
	if err != nil {
		return t, err
	}

	var tags TagsResponse
	err = json.Unmarshal(body, &tags)
	if err != nil {
		return t, err
	}

	tagsWithTimeAndArchitecture := make([]struct {
		Tag          string `json:"tag"`
		CreatedTime  string `json:"createdTime"`
		Architecture string `json:"architecture"`
	}, len(tags.Tags))

	// 获取每个标签的创建时间和架构信息并存储到 tagsWithTimeAndArchitecture 结构体中
	for i, tag := range tags.Tags {
		createdTime, err := getCreatedTime(repo, tag, registryURL, httpClient)
		if err != nil {
			return t, err
		}

		architectures, err := getArchitectures(repo, tag, registryURL, httpClient)
		if err != nil {
			return t, err
		}
		architecture := strings.Join(architectures, ",")

		tagsWithTimeAndArchitecture[i] = struct {
			Tag          string `json:"tag"`
			CreatedTime  string `json:"createdTime"`
			Architecture string `json:"architecture"`
		}{Tag: tag, CreatedTime: createdTime, Architecture: architecture}
	}

	// 根据创建时间排序
	sort.Slice(tagsWithTimeAndArchitecture, func(i, j int) bool {
		time1, _ := time.Parse(time.RFC3339Nano, tagsWithTimeAndArchitecture[i].CreatedTime)
		time2, _ := time.Parse(time.RFC3339Nano, tagsWithTimeAndArchitecture[j].CreatedTime)
		return time1.After(time2)
	})

	// 根据需求截取展示的 tags
	if !showAll && len(tagsWithTimeAndArchitecture) > 5 {
		tagsWithTimeAndArchitecture = tagsWithTimeAndArchitecture[:5]
	}

	for _, tagsWithTimeAndArchitecture := range tagsWithTimeAndArchitecture {
		tag := tagsWithTimeAndArchitecture.Tag
		createdTime := tagsWithTimeAndArchitecture.CreatedTime
		architecture := tagsWithTimeAndArchitecture.Architecture

		// 解析时间字符串为时间对象
		time, err := time.Parse(time.RFC3339Nano, createdTime)
		if err != nil {
			fmt.Println("解析时间出错:", err)
		}
		// 格式化为常见的时间样式
		formattedTime := time.Format("2006-01-02 15:04:05")

		row := table.Row{repo, tag, architecture, formattedTime}
		t.AppendRow(row)
	}
	return t, err
}

func getCreatedTime(repo, tag, registryURL string, httpClient *http.Client) (string, error) {
	// 创建一个带有自定义标头的 HTTP 请求，获取镜像的清单
	manifestURL := fmt.Sprintf("%s/v2/%s/manifests/%s", registryURL, repo, tag)

	manifestRequest, err := http.NewRequest("GET", manifestURL, nil)
	if err != nil {
		return "", err
	}
	manifestRequest.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	// 发送请求并获取响应
	manifestResponse, err := httpClient.Do(manifestRequest)
	if err != nil {
		return "", err
	}
	defer manifestResponse.Body.Close()

	body, err := ioutil.ReadAll(manifestResponse.Body)
	if err != nil {
		return "", err
	}

	var manifest ManifestResponse
	err = json.Unmarshal(body, &manifest)
	if err != nil {
		return "", err
	}

	configDigest := manifest.Config.Digest

	// 创建一个带有自定义标头的 HTTP 请求，获取配置对象的内容
	configBlobURL := fmt.Sprintf("%s/v2/%s/blobs/%s", registryURL, repo, configDigest)
	configBlobRequest, err := http.NewRequest("GET", configBlobURL, nil)
	if err != nil {
		return "", err
	}
	configBlobRequest.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	// 发送请求并获取响应
	configBlobResponse, err := httpClient.Do(configBlobRequest)
	if err != nil {
		return "", err
	}
	defer configBlobResponse.Body.Close()

	body, err = ioutil.ReadAll(configBlobResponse.Body)
	if err != nil {
		return "", err
	}

	var configBlob ConfigBlob
	err = json.Unmarshal(body, &configBlob)
	if err != nil {
		return "", err
	}

	return configBlob.Created, nil
}

func getArchitectures(repo, tag, registryURL string, httpClient *http.Client) ([]string, error) {

	manifestURL := fmt.Sprintf("%s/v2/%s/manifests/%s", registryURL, repo, tag)

	// 创建一个不验证证书的安全传输
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// 创建 HTTP 客户端
	client := &http.Client{Transport: tr}

	// 创建 GET 请求
	req, err := http.NewRequest("GET", manifestURL, nil)
	if err != nil {
		log.Fatal("创建请求失败:", err)
		return []string{}, err
	}

	// 设置请求头部
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("请求失败:", err)
		return []string{}, err
	}
	defer resp.Body.Close()

	// 解析 JSON
	var result ManifestList
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatal("解析 JSON 失败:", err)
		return []string{}, err
	}

	// 检查响应类型并输出架构信息
	var res []string
	if len(result.Manifests) > 0 {
		for _, manifest := range result.Manifests {
			res = append(res, manifest.Platform.Architecture)
		}
	} else {
		res = append(res, result.Architecture)
	}
	return res, err
}
