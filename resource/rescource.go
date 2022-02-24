package resource

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

var resourceDir = "res"

func SetResourceDir(dir string) {
	resourceDir = dir
}

type ResourceManager struct {
	basePath      string
	requestHeader http.Header
}

// 新建资源管理器，basePath为资源文件夹内的子目录
func NewResourceManager(basePath string) *ResourceManager {
	basePath = filepath.Join(resourceDir, basePath)

	return &ResourceManager{
		basePath: basePath,
	}
}

func (rm *ResourceManager) SetRequestHeader(header http.Header) {
	rm.requestHeader = header
}

func (rm *ResourceManager) Load(pathToFile string) (res Resource, err error) {
	relPath := filepath.Join(rm.basePath, pathToFile)
	file, err := os.Open(relPath)
	if err != nil {
		return
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return
	}
	absPath, err := filepath.Abs(pathToFile)
	if err != nil {
		return
	}

	res = Resource{
		name:    path.Base(absPath),
		ext:     filepath.Ext(pathToFile),
		relPath: relPath,
		absPath: absPath,
		content: bytes,
	}
	return
}

func (rm *ResourceManager) LoadImage(pathToFile string) (res ImageResource, err error) {
	r, err := rm.Load(pathToFile)
	if err != nil {
		return
	}
	res = ImageResource{
		Resource: r,
	}
	return
}

// 获取网络资源。
//
// save为true时，将资源保存到本地的pathToFile中。若pathToFile为空，则保存到默认路径（network文件夹下）。
// 如果save为false，则忽略pathToFile。
func (rm *ResourceManager) LoadNetworkResource(url string, save bool, pathToFile string) (res NetworkResource, err error) {
	// 创建http请求并设置请求头
	httpClient := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	if rm.requestHeader != nil {
		req.Header = rm.requestHeader
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	res.mimeType = resp.Header.Get("Content-Type")
	res.netUrl = url
	res.content = bytes
	res.name = path.Base(url)
	res.ext = filepath.Ext(url)

	if save {
		// 默认保存到network文件夹下，文件名为url的文件名。
		if pathToFile == "" {
			pathToFile = filepath.Join("network", res.name)
		}
		err = rm.Save(pathToFile, bytes)
		if err != nil {
			return
		}
		relPath := filepath.Join(rm.basePath, pathToFile)
		absPath, _ := filepath.Abs(relPath)
		res.relPath = relPath
		res.absPath = absPath
		res.name = path.Base(relPath)
		res.ext = filepath.Ext(relPath)
	}

	return
}

func (rm *ResourceManager) Save(pathToFile string, content []byte) (err error) {
	relPath := filepath.Join(rm.basePath, pathToFile)
	// 防止目录不存在，先创建目录
	err = os.MkdirAll(filepath.Dir(relPath), 0755)
	if err != nil {
		return
	}
	err = os.WriteFile(relPath, content, 0644)
	return
}

func (rm *ResourceManager) Delete(pathToFile string) (err error) {
	relPath := filepath.Join(rm.basePath, pathToFile)
	err = os.Remove(relPath)
	return
}

type Resource struct {
	name    string
	ext     string
	relPath string
	absPath string
	content []byte
}

// 文件名，带有扩展名
func (r Resource) Name() string {
	return r.name
}

// 扩展名，以点开头
func (r Resource) Ext() string {
	return r.ext
}

// 相对路径，相对于工作目录
func (r Resource) RelPath() string {
	return r.relPath
}

// 绝对路径
func (r Resource) AbsPath() string {
	return r.absPath
}

// 资源的内容
func (r Resource) Content() []byte {
	return r.content
}

// 文件Uri
func (r Resource) FileUri() string {
	return fmt.Sprintf("file://%s", r.AbsPath())
}

// 图片资源
type ImageResource struct {
	Resource
}

// data uri
func (r ImageResource) Base64Uri() string {
	return fmt.Sprintf("data:image/%s;base64,%s", r.Ext(), base64.StdEncoding.EncodeToString(r.Content()))
}

// 网络资源。当资源未保存到本地时，relPath、absPath为空。
type NetworkResource struct {
	Resource
	netUrl   string
	mimeType string
}

// 网络资源的url
func (r NetworkResource) NetUrl() string {
	return r.netUrl
}

func (r NetworkResource) MimeType() string {
	return r.mimeType
}

// data uri
func (r NetworkResource) Base64Uri() string {
	return fmt.Sprintf("data:%s;base64,%s", r.MimeType(), base64.StdEncoding.EncodeToString(r.Content()))
}
