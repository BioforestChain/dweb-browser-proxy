package main

import (
	"context"
	"fmt"
	timeHelper "github.com/BioforestChain/dweb-browser-proxy/internal/pkg/util/time"
	ipc2 "github.com/BioforestChain/dweb-browser-proxy/pkg/ipc"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	//url := "ws://127.0.0.1:8000/ws?client_id=f544391c9215a6bfbb6f573b25c40d38"
	url := "ws://127.0.0.1:8000/proxy/ws?client_id=127.0.0.1"
	//url := "ws://127.0.0.1:8000/proxy/ws"
	conn, res, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Println("Error connecting to server: ", err)
	}
	defer conn.Close()

	log.Println("Connected to WebSocket server", res.Status)

	ipcConn := newIPCConn(conn)

	for {
		//四个字节：uin32是长度,取多少长度的内容
		//O  {"req_id":1,"method":"GET","url":"http://127.0.0.1:8000/ipc/test?client_id=127.0.0.2","headers":{"Accept":"*/*","Accept-Encoding":"gzip, deflate, br","Connection":"keep-alive","User-Agent":"Apifox/1.0.0 (https://apifox.com)"},"metaBody":{"type":5,"senderUid":7,"receiverUid":0,"data":"","streamId":"","metaId":"1xyf90kyQe8="},"type":0}
		//O  {"req_id":3,"method":"GET","url":"http://127.0.0.1:8000/ipc/test?client_id=127.0.0.1","headers":{"Accept":"*/*","Accept-Encoding":"gzip, deflate, br","Connection":"keep-alive","User-Agent":"Apifox/1.0.0 (https://apifox.com)"},"metaBody":{"type":5,"senderUid":7,"receiverUid":0,"data":"","streamId":"","metaId":"eK+VxCgZxng="},"type":0}

		//�  {"req_id":1,"method":"GET","url":"http://127.0.0.1:8000/ipc/test?client_id=f544391c9215a6bfbb6f573b25c40d38","headers":{"Accept":"*/*","Accept-Encoding":"gzip, deflate, br","Connection":"keep-alive","Content-Length":"173","Content-Type":"multipart/form-data; boundary=--------------------------987399823725143514993580","User-Agent":"Apifox/1.0.0 (https://apifox.com)"},"metaBody":{"type":5,"senderUid":4,"receiverUid":0,"data":"LS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLTk4NzM5OTgyMzcyNTE0MzUxNDk5MzU4MA0KQ29udGVudC1EaXNwb3NpdGlvbjogZm9ybS1kYXRhOyBuYW1lPSJjbGllbnRfaWQiDQoNCjEyNy4wLjAuMQ0KLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLS0tLTk4NzM5OTgyMzcyNTE0MzUxNDk5MzU4MC0tDQo=","streamId":"","metaId":"bzkzZyhdh2M="},"type":0}
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message: ", err)
			return
		}
		//fmt.Printf("Received message: %s\n", message)

		if err := ipcConn.inputStream.Enqueue(message); err != nil {
			log.Println("enqueue err: ", err)
		}
	}

}

type IPCConn struct {
	ipc         ipc2.IPC
	inputStream *ipc2.ReadableStream
	conn        *websocket.Conn
}

// 获取文件名
func getFileName(part string) string {
	// 使用正则表达式等方式获取文件名
	re := regexp.MustCompile(`filename="(.*)"`)
	subs := re.FindStringSubmatch(part)
	if len(subs) > 1 {
		return strconv.Itoa(int(timeHelper.Time())) + `_` + subs[1]
	}
	return ""
}

// 获取文件数据
func getDataBetween(part string, start, end string) string {
	// 返回part中start和end之间的数据
	s := strings.Index(part, start)
	if s < 0 {
		return ""
	}
	s += len(start)
	e := strings.Index(part[s:], end)
	if e < 0 {
		return ""
	}
	return part[s : s+e]
}

// 保存文件
func saveFile(name, data string) {
	// 保存文件到指定目录
	filepath.Ext(name)

	//img, _ := os.Create(filepath.Join("/tmp", name))
	proPath := ProjectPath()
	fmt.Printf("proPath: %#v\n", proPath)
	img, _ := os.Create(filepath.Join(proPath+"server/tmp", name))
	img.Write([]byte(data))
	img.Close()
}

func newIPCConn(conn *websocket.Conn) *IPCConn {
	serverIPC := ipc2.NewReadableStreamIPC(ipc2.SERVER, ipc2.SupportProtocol{
		Raw:         true,
		MessagePack: false,
		ProtoBuf:    false,
	})

	// 监听并处理请求，echo请求数据
	serverIPC.OnRequest(func(req interface{}, ic ipc2.IPC) {
		request := req.(*ipc2.Request)
		log.Println("on request: ", request.ID)

		url, _ := url.ParseRequestURI(request.URL)
		if (url.Host+url.Path) == "127.0.0.1:8000/ipc/test" && request.Method == "GET" {
			//bodyReceiver := request.Body.(*ipc.BodyReceiver)
			//body := bodyReceiver.GetMetaBody().Data
			//log.Println("onRequest: ", request.URL, string(body), ic)

			body := `{"code": 0, "message": "hi"}`
			res := ipc2.NewResponse(
				request.ID,
				200,
				ipc2.NewHeaderWithExtra(map[string]string{
					"Content-Type": "application/json",
				}),
				ipc2.NewBodySender([]byte(body), ic),
				ic,
			)

			if err := ic.PostMessage(context.TODO(), res); err != nil {
				log.Println("post message err: ", err)
			}
		}
	})

	// 监听并处理请求，echo请求数据
	serverIPC.OnRequest(func(req interface{}, ic ipc2.IPC) {
		request := req.(*ipc2.Request)
		log.Println("on request: ", request.ID)

		url, _ := url.ParseRequestURI(request.URL)

		if (url.Host + url.Path) == "127.0.0.1:8000/ipc/stream" {
			var res *ipc2.Response
			path, _ := getAbsPath("server/golang.png")
			f, err := os.Open(path)
			if err != nil {
				res = ipc2.FromResponseText(
					request.ID,
					http.StatusInternalServerError,
					ipc2.NewHeader(),
					http.StatusText(http.StatusInternalServerError),
					ic,
				)
			} else {
				bodyStream := ipc2.NewReadableStream()

				var total int
				info, _ := f.Stat()
				log.Println(info.Size())

				go func() {
					for {
						data := make([]byte, 1024)
						n, err := f.Read(data)
						if err != nil {
							bodyStream.Controller.Close()
							break
						}
						total += n

						_ = bodyStream.Enqueue(data[:n])
						log.Println("total: ", total)
					}

					log.Println(f.Close())
				}()

				res = ipc2.FromResponseStream(
					request.ID,
					http.StatusOK,
					ipc2.NewHeaderWithExtra(map[string]string{
						"Content-Type": "image/png",
					}),
					bodyStream,
					ic,
				)
			}

			if err := ic.PostMessage(context.TODO(), res); err != nil {
				log.Println("post message err: ", err)
			}
		}
	})
	// POST http test
	serverIPC.OnRequest(func(req interface{}, ic ipc2.IPC) {
		request := req.(*ipc2.Request)

		log.Println("on request: ", request.ID)
		bodyReceiver := request.Body.(*ipc2.BodyReceiver)
		body := bodyReceiver.GetMetaBody().Data

		//----------------------------716381082216618958260243
		//Content-Disposition: form-data; name="client_id"
		//
		//f544391c9215a6bfbb6f573b25c40d38
		//----------------------------716381082216618958260243
		//Content-Disposition: form-data; name="file_name"; filename="background.75d5944.png"
		//Content-Type: image/png
		//
		//�PNG
		//...
		//etc
		//----------------------------716381082216618958260243--
		url, _ := url.ParseRequestURI(request.URL)
		//upload
		//file_name
		if (url.Host+url.Path) == "127.0.0.1:8000/ipc/test" && request.Method == "POST" {
			//log.Println("onRequest: ", string(body))
			//log.Println("onRequest len is: ", len(string(body)))
			// 分割请求体
			// 找到第一个boundary位置
			firstIdx := strings.Index(string(body), "\r\n--")

			// 从第一个boundary之后找结尾boundary
			secondIdx := strings.Index(string(body)[firstIdx+4:], "\r\n--")

			// 计算第二个boundary在整个body中的位置
			secondIdx += firstIdx + 4
			// 第一段从开始到第一个boundary
			//firstPart := string(body[:firstIdx])

			// 第二段从第一个boundary到第二个boundary
			secondPart := string(body[firstIdx:secondIdx])
			//log.Printf("firstPart body is: %#v\n", firstPart)
			//log.Printf("secondPart body is: %#v\n", secondPart)
			//filename
			// 再根据文件名分割
			nameIdx := strings.Index(secondPart, "\r\n\r\n")

			nameSection := secondPart[:nameIdx]
			fileSection := secondPart[nameIdx+4:]
			filename := getFileName(nameSection)
			// 获取文件内容
			//log.Printf("filename  is: %#v\n", filename)
			//log.Printf("fileSection  is: %#v\n", fileSection)
			//保存文件
			saveFile(filename, fileSection)

			bodyOther := `{"code": 0, "message": "hi by post"}`

			res := ipc2.NewResponse(
				request.ID,
				200,
				ipc2.NewHeaderWithExtra(map[string]string{
					"Content-Type": "application/json",
				}),
				ipc2.NewBodySender([]byte(bodyOther), ic),
				ic,
			)

			if err := ic.PostMessage(context.TODO(), res); err != nil {
				log.Println("post message err: ", err)
			}
		}
	})

	///proxy/pubsub/subscribe_msg
	serverIPC.OnRequest(func(req interface{}, ic ipc2.IPC) {

		request := req.(*ipc2.Request)
		log.Println("on request: ", request.ID)
		url, _ := url.ParseRequestURI(request.URL)

		if (url.Host+url.Path) == "127.0.0.1:8000/proxy/pubsub/publish_msg" ||
			(url.Host+url.Path) == "127.0.0.1:8000/proxy/pubsub/subscribe_msg" && request.Method == "POST" {

			bodyReceiver := request.Body.(*ipc2.BodyReceiver)
			body1 := bodyReceiver.GetMetaBody().Data
			log.Println("onRequest: ", request.URL, string(body1), ic)
			log.Printf("subscribe_msg post body is: %#v\n", request.Body)
			log.Printf("subscribe_msg Header is %#v\n: ", request.Header)

			body := `{"code": 0, "message": "subscribe_msg"}`
			res := ipc2.NewResponse(
				request.ID,
				200,
				ipc2.NewHeaderWithExtra(map[string]string{
					"Content-Type": "application/json",
				}),
				ipc2.NewBodySender([]byte(body), ic),
				ic,
			)

			if err := ic.PostMessage(context.TODO(), res); err != nil {
				log.Println("post message err: ", err)
			}
		}
	})

	inputStream := ipc2.NewReadableStream()

	ipcConn := &IPCConn{
		ipc:         serverIPC,
		inputStream: inputStream,
		conn:        conn,
	}

	go func() {
		defer ipcConn.Close()
		// 读取inputStream数据并emit消息（接收消息并处理，然后把结果发送至内部流）
		if err := serverIPC.BindInputStream(inputStream); err != nil {
			panic(err)
		}
	}()

	go func() {
		// 读取内部流数据，然后response
		serverIPC.ReadOutputStream(func(data []byte) {
			if err := ipcConn.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				log.Println("res msg err: ", err)
				panic(err)
			}
		})
	}()
	fmt.Printf("ipcConn: %#v\n", ipcConn)
	return ipcConn
}

func (i *IPCConn) Close() {
	i.ipc.Close()
	i.inputStream.Controller.Close()
	i.conn.Close()
}

func getAbsPath(path string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(cwd, path), nil
}
