package main

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	timeHelper "proxyServer/internal/helper/time"
	"proxyServer/internal/packed"
	"proxyServer/ipc"
)

func main() {
	//url := "ws://127.0.0.1:8000/ws?client_id=f544391c9215a6bfbb6f573b25c40d38"
	url := "ws://127.0.0.1:8000/ws?client_id=127.0.0.1"
	conn, res, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Println("Error connecting to server: ", err)
	}
	defer conn.Close()

	log.Println("Connected to WebSocket server", res.Status)

	ipcConn := newIPCConn(conn)

	for {
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
	ipc         ipc.IPC
	inputStream *ipc.ReadableStream
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
	proPath := packed.ProjectPath()
	fmt.Printf("proPath: %#v\n", proPath)
	img, _ := os.Create(filepath.Join(proPath+"server/tmp", name))
	img.Write([]byte(data))
	img.Close()
}

func newIPCConn(conn *websocket.Conn) *IPCConn {
	serverIPC := ipc.NewReadableStreamIPC(ipc.SERVER, ipc.SupportProtocol{
		Raw:         true,
		MessagePack: false,
		ProtoBuf:    false,
	})

	// 监听并处理请求，echo请求数据
	serverIPC.OnRequest(func(req interface{}, ic ipc.IPC) {
		request := req.(*ipc.Request)
		log.Println("on request: ", request.ID)

		url, _ := url.ParseRequestURI(request.URL)
		if (url.Host+url.Path) == "127.0.0.1:8000/ipc/test" && request.Method == "GET" {
			//bodyReceiver := request.Body.(*ipc.BodyReceiver)
			//body := bodyReceiver.GetMetaBody().Data
			//log.Println("onRequest: ", request.URL, string(body), ic)

			body := `{"code": 0, "message": "hi"}`
			res := ipc.NewResponse(
				request.ID,
				200,
				ipc.NewHeaderWithExtra(map[string]string{
					"Content-Type": "application/json",
				}),
				ipc.NewBodySender([]byte(body), ic),
				ic,
			)

			if err := ic.PostMessage(context.TODO(), res); err != nil {
				log.Println("post message err: ", err)
			}
		}
	})

	// 监听并处理请求，echo请求数据
	serverIPC.OnRequest(func(req interface{}, ic ipc.IPC) {
		request := req.(*ipc.Request)
		log.Println("on request: ", request.ID)

		url, _ := url.ParseRequestURI(request.URL)

		if (url.Host + url.Path) == "127.0.0.1:8000/ipc/stream" {
			var res *ipc.Response
			path, _ := getAbsPath("server/golang.png")
			f, err := os.Open(path)
			if err != nil {
				res = ipc.FromResponseText(
					request.ID,
					http.StatusInternalServerError,
					ipc.NewHeader(),
					http.StatusText(http.StatusInternalServerError),
					ic,
				)
			} else {
				bodyStream := ipc.NewReadableStream()

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

				res = ipc.FromResponseStream(
					request.ID,
					http.StatusOK,
					ipc.NewHeaderWithExtra(map[string]string{
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
	serverIPC.OnRequest(func(req interface{}, ic ipc.IPC) {
		request := req.(*ipc.Request)

		log.Println("on request: ", request.ID)
		bodyReceiver := request.Body.(*ipc.BodyReceiver)
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

			res := ipc.NewResponse(
				request.ID,
				200,
				ipc.NewHeaderWithExtra(map[string]string{
					"Content-Type": "application/json",
				}),
				ipc.NewBodySender([]byte(bodyOther), ic),
				ic,
			)

			if err := ic.PostMessage(context.TODO(), res); err != nil {
				log.Println("post message err: ", err)
			}
		}
	})

	///proxy/pubsub/subscribe_msg
	serverIPC.OnRequest(func(req interface{}, ic ipc.IPC) {

		request := req.(*ipc.Request)
		log.Println("on request: ", request.ID)
		url, _ := url.ParseRequestURI(request.URL)

		if (url.Host+url.Path) == "127.0.0.1:8000/proxy/pubsub/publish_msg" || (url.Host+url.Path) == "127.0.0.1:8000/proxy/pubsub/subscribe_msg" && request.Method == "POST" {

			bodyReceiver := request.Body.(*ipc.BodyReceiver)
			body1 := bodyReceiver.GetMetaBody().Data
			log.Println("onRequest: ", request.URL, string(body1), ic)
			log.Printf("subscribe_msg post body is: %#v\n", request.Body)
			log.Printf("subscribe_msg Header is %#v\n: ", request.Header)

			body := `{"code": 0, "message": "subscribe_msg"}`
			res := ipc.NewResponse(
				request.ID,
				200,
				ipc.NewHeaderWithExtra(map[string]string{
					"Content-Type": "application/json",
				}),
				ipc.NewBodySender([]byte(body), ic),
				ic,
			)

			if err := ic.PostMessage(context.TODO(), res); err != nil {
				log.Println("post message err: ", err)
			}
		}
	})

	inputStream := ipc.NewReadableStream()

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
