package CSbotToGo

import "C"
import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	pbx "github.com/tinode/chat/pbx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//bot.ServerPresEvent = func(sender interface{}, e ServerPresEventArgs) {
//	// реализация обработчика
//}
//
//// вызов события ServerPresEvent
//bot.ServerPresEvent(bot, ServerPresEventArgs{})

type ChatBot struct {
	UploadedAttachmentInfo
	ServerPresEventArgs
	ServerMetaEventArgs
	CtrlMessageEventArgs
	ServerDataEventArgs
	Subscriber
	ChatBotPlugin
	Future
	FutureTypes

	AppName          string
	AppVersion       string
	LibVersion       string
	Platform         string
	UploadEndpoint   string
	DownloadEndpoint string
	BotUID           string
	NextTid          int64
	ServerHost       string
	Listen           string
	CookieFile       string
	Schema           string
	Secret           string
	Subscribers      map[string]Subscriber
	Server           *grpc.Server
	//458
	//AsyncDuplexStreamingCall<ClientMsg, ServerMsg> client;
	//Channel channel;
	//CancellationTokenSource cancellationTokenSource;
	//Queue<ClientMsg> sendMsgQueue;
	BotResponse             IBotResponse
	Client                  pbx.Node_MessageLoopClient
	Cl2                     pbx.NodeClient
	Cl3                     pbx.Node_MessageLoopServer
	Channel                 grpc.ClientConnInterface
	CancellationTokenSource context.CancelFunc
	SendMsgQueue            []*pbx.ClientMsg
	Subscriptions           map[string]bool
	OnCompletion            map[string]Future
	Token                   string
	ApiKey                  string
	ApiBaseUrl              string
	// func chatbot
	// 543
	ServerDataEvent   func(sender interface{}, e *ServerDataEventArgs)
	ServerMetaEvent   func(sender interface{}, e *ServerMetaEventArgs)
	ServerPresEvent   func(sender interface{}, e *ServerPresEventArgs)
	CtrlMessageEvent  func(sender interface{}, e CtrlMessageEventArgs)
	LoginSuccessEvent func(sender interface{}, e EventArgs)
	LoginFailedEvent  func(sender interface{}, e EventArgs)
	DisconnectedEvent chan bool
}

func (c *ChatBot) SendMessageLoop(ctx context.Context) {
	log.Println("Message send queue started...")
	if len(c.SendMsgQueue) > 0 {
		msg := c.SendMsgQueue[0]
		c.SendMsgQueue = c.SendMsgQueue[1:]

		err := c.Client.SendMsg(msg)
		//err := c.Client.RecvMsg(msg)

		if err != nil {
			log.Printf("Send Message Error: %v, Failed message will be put back to queue...\n", err)
			c.SendMsgQueue[0] = msg
			time.Sleep(1 * time.Second)
		}
	} else {
		time.Sleep(10 * time.Millisecond)
	}
	if len(c.SendMsgQueue) > 0 {
		msg := c.SendMsgQueue[0]
		c.SendMsgQueue = c.SendMsgQueue[1:]

		err := c.Client.SendMsg(msg)
		if err != nil {
			log.Printf("Send Message Error: %v, Failed message will be put back to queue...\n", err)
			c.SendMsgQueue[0] = msg
			time.Sleep(1 * time.Second)
		}
	} else {
		time.Sleep(10 * time.Millisecond)
	}
	if len(c.SendMsgQueue) > 0 {
		msg := c.SendMsgQueue[0]
		c.SendMsgQueue = c.SendMsgQueue[1:]

		err := c.Client.SendMsg(msg)
		if err != nil {
			log.Printf("Send Message Error: %v, Failed message will be put back to queue...\n", err)
			c.SendMsgQueue[0] = msg
			time.Sleep(1 * time.Second)
		}
	} else {
		time.Sleep(10 * time.Millisecond)
	}
	if len(c.SendMsgQueue) > 0 {
		msg := c.SendMsgQueue[0]
		c.SendMsgQueue = c.SendMsgQueue[1:]

		err := c.Client.SendMsg(msg)
		if err != nil {
			log.Printf("Send Message Error: %v, Failed message will be put back to queue...\n", err)
			c.SendMsgQueue[0] = msg
			time.Sleep(1 * time.Second)
		}
	} else {
		time.Sleep(10 * time.Millisecond)
	}

	log.Println("Detect cancel message, stop sending message...")
	//for {
	//	select {
	//	case <-ctx.Done():
	//		log.Println("Detect cancel message, stop sending message...")
	//		return
	//	default:
	//		if len(c.SendMsgQueue) > 0 {
	//			msg := &c.SendMsgQueue[0]
	//			c.SendMsgQueue = c.SendMsgQueue[1:]
	//
	//			err := c.Client.SendMsg(msg)
	//			if err != nil {
	//				log.Printf("Send Message Error: %v, Failed message will be put back to queue...\n", err)
	//				c.SendMsgQueue[0] = *msg
	//				time.Sleep(1 * time.Second)
	//			}
	//		} else {
	//			time.Sleep(10 * time.Millisecond)
	//		}
	//	}
	//}
}
func (c *ChatBot) DelSubscription(topic string) {
	if _, ok := c.Subscriptions[topic]; ok {
		delete(c.Subscriptions, topic)
	}
}
func (c *ChatBot) AddFuture(tid string, bundle Future) {
	c.OnCompletion[tid] = bundle
}
func (c *ChatBot) GetNextTid() string {
	c.NextTid += 1
	return strconv.Itoa(int(c.NextTid))
}
func (c *ChatBot) Log(title, message string) {
	fmt.Printf("[%s] - [%s] : %s\n", time.Now().Format("2006-01-02 15:04:05"), title, message)
}
func (c *ChatBot) SetHttpApi(apiBaseUtl, apiKey string) {
	c.ApiBaseUrl = apiBaseUtl
	c.ApiKey = apiKey
}
func (c *ChatBot) ReadAuthCookie() (schema string, secret []byte, ok bool) {
	if _, err := os.Stat(c.CookieFile); os.IsNotExist(err) {
		return "", nil, false
	}

	f, err := os.Open(c.CookieFile)
	if err != nil {
		return "", nil, false
	}
	defer f.Close()

	cookies := make(map[string]string)
	if err := json.NewDecoder(f).Decode(&cookies); err != nil {
		return "", nil, false
	}

	schema = cookies["schema"]
	if schema == "" {
		return "", nil, false
	}

	secretStr := cookies["secret"]
	if secretStr == "" {
		return "", nil, false
	}

	if schema == "token" {
		secretBytes, err := base64.StdEncoding.DecodeString(string(secretStr))
		if err != nil {
			return "", nil, false
		}
		secret = secretBytes
	} else {
		secret = []byte(secretStr)
	}

	return schema, secret, true
}
func (c *ChatBot) OnServerPresEvent(e *ServerPresEventArgs) {
	if c.ServerPresEvent != nil {
		c.ServerPresEvent(c.ServerPresEvent, e)
	}
	//c.ServerPresEvent(c, e)
}
func (c *ChatBot) OnServerMetaEvent(e *ServerMetaEventArgs) {
	if c.ServerMetaEvent != nil {
		c.ServerMetaEvent(c, e)
	}
}
func (c *ChatBot) OnGetMeta(meta *pbx.ServerMeta) {
	if meta.Sub != nil {
		for _, sub := range meta.Sub {
			userId := sub.UserId
			online := sub.Online
			topic := sub.Topic
			publicInfo := string(sub.Public)
			var subObj map[string]interface{}
			if err := json.Unmarshal([]byte(publicInfo), &subObj); err != nil {
				panic(err)
			}
			userName := topic
			typeStr := "group"
			if subObj != nil {
				typeStr = "user"
			}
			photoData := ""
			photoType := ""
			if subObj != nil {
				userName = fmt.Sprintf("%v", subObj["fn"])
				if _, ok := subObj["photo"]; ok {
					photoData = subObj["photo"].(map[string]interface{})["data"].(string)
					photoType = subObj["photo"].(map[string]interface{})["type"].(string)
				}
			}
			c.AddSubscriber(Subscriber{
				MessageBase: MessageBase{},
				UserId:      userId,
				Topic:       topic,
				UserName:    userName,
				PhotoData:   photoData,
				Type:        typeStr,
				Online:      online,
				PhotoType:   photoType,
			})

		}
	}
}
func (c *ChatBot) OnDisconnected() {
	select {
	case c.DisconnectedEvent <- true:
	default:
	}
}
func (c *ChatBot) ClientReset() {
	c.Token = ""
	for k := range c.Subscriptions {
		delete(c.Subscriptions, k)
	}
	for k := range c.OnCompletion {
		delete(c.OnCompletion, k)
	}
	for k := range c.Subscribers {
		delete(c.Subscribers, k)
	}
	c.SendMsgQueue = []*pbx.ClientMsg{}
}
func (c *ChatBot) Start() {
	fmt.Println("init Server")
	c.Server = c.InitServer()

	c.Client = c.InitClient()
	ctx, _ := context.WithCancel(context.Background())
	c.SendMessageLoop(context.Background())
	err := c.ClientMessageLoop(ctx)
	if err != nil {
		go func() {
			for {
				select {
				case <-c.DisconnectedEvent:
					c.Log("Connection Broken", fmt.Sprintf("Connection Closed: %s", err))
					time.Sleep(2 * time.Second)
					c.ClientReset()
					c.Client = c.InitClient()
				default:

				}
			}
		}()
	}
	//go func() {
	//	for {
	//		select {
	//		case <-ctx.Done():
	//			// контекст был отменен, выходим из цикла
	//			return
	//
	//		default:
	//			err := c.ClientMessageLoop(ctx)
	//			if err != nil {
	//				go func() {
	//					for {
	//						select {
	//						case <-c.DisconnectedEvent:
	//							c.Log("Connection Broken", fmt.Sprintf("Connection Closed: %s", err))
	//							time.Sleep(2 * time.Second)
	//							c.ClientReset()
	//							c.Client = c.InitClient()
	//						default:
	//
	//						}
	//					}
	//				}()
	//			}
	//		}
	//	}
	//}()
}

//ClientMessageLoop  остановился на этой функции
func (c *ChatBot) ClientMessageLoop(ctx context.Context) error {

	var err error
	for {
		select {
		case <-ctx.Done():
			return err
		default:
			for {
				//pbx.Node_MessageLoopClient()
				//response, err := c.Client.Recv()
				//resp, err := c.Cl3.Recv()
				response, err := c.Client.Recv()
				//err = c.Client.SendMsg(&pbx.ClientMsg{Message: &pbx.ClientMsg_Note{Note: &pbx.ClientNote{
				//	Topic:   response.Topic,
				//	What:    0,
				//	SeqId:   0,
				//	Unread:  0,
				//	Event:   0,
				//	Payload: nil,
				//}}})
				//	err = c.Client.RecvMsg(proto.Message(response))
				//fmt.Println(resp.Message, resp.GetMessage())

				if err == io.EOF {
					break
				}
				if err != nil {
					log.Fatalf("Error while receiving server response: %v", err)
				}
				//fmt.Println(response.GetData(), c.Data)
				fmt.Println(response.GetData(), " ", response.GetPres())

				if response.GetCtrl() != nil {
					messageh := fmt.Sprintf("ID=%v  Code=%v  Text=%v  Params=%v\n", response.GetCtrl().Id, response.GetCtrl().Code, response.GetCtrl().Text, response.GetCtrl().Params)
					c.Log("Ctrl Message", messageh)
					c.ExecFuture(response.GetCtrl().Id, int(response.GetCtrl().Code), response.GetCtrl().Text, response.GetCtrl().Topic, response.GetCtrl().Params)
				} else if response.GetData() != nil {
					c.OnServerDataEvent(c.ServerDataEventArgs.ServerDataEventArgs(response.GetData()))
					if response.GetData().FromUserId != c.BotUID {
						c.ClientPost(c.NoteRead(response.GetData().Topic, int(response.GetData().SeqId)))
						clone := response.GetData()
						time.Sleep(50 * time.Millisecond)
						if c.BotResponse != nil {
							reply, _ := c.BotResponse.ThinkAndReply(clone)
							if reply != nil {
								c.ClientPost(c.Publish(response.GetData().Topic, reply))
							}
						} else {
							c.ClientPost(c.Publish(response.GetData().Topic, &ChatMessage{
								MessageBase: MessageBase{},
								Text:        "I don't know how to talk with you, maybe my father didn't put my brain in my head...",
								Fmt:         nil,
								Ent:         nil,
								IsPlainText: false,
								MessageType: "",
							}))
						}
					}
				} else if response.GetPres() != nil {
					if response.GetPres().Topic == "me" {
						if _, ok := c.Subscriptions[response.GetPres().Src]; !ok {
							if response.GetPres().What == pbx.ServerPres_ON || response.GetPres().What == pbx.ServerPres_MSG {
								c.ClientPost(c.Subscribe(response.GetPres().Src))
							} else if ok {
								if response.GetPres().What == pbx.ServerPres_OFF {
									c.ClientPost(c.Leave(response.GetPres().Src))
								}
							}
						}
						c.OnServerPresEvent(c.ServerPresEventArgs.ServerPresEventArgs(response.GetPres()))

					}
				} else if response.GetMeta() != nil {
					//c.OnServerMetaEvent(c.ServerMetaEventArgs.ServerMetaEventArgs(response.GetMeta()))
				}
			}
		}
	}
	return err
}
func (c *ChatBot) Leave(topic string) *pbx.ClientMsg {
	tid := c.GetNextTid()
	c.AddFuture(tid, c.Future.Future(tid, Leave, func(topicName string, unused map[string][]byte) {
		c.DelSubscription(topicName)
	}, ""))
	lve := &pbx.ClientMsg_Leave{Leave: &pbx.ClientLeave{
		Id:    tid,
		Topic: topic,
		Unsub: false,
	}}
	return &pbx.ClientMsg{
		Message: lve,
		Extra:   nil,
	}
}
func (c *ChatBot) Publish(topic string, msg *ChatMessage) *pbx.ClientMsg {
	tid := c.GetNextTid()
	bytes, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	pub := &pbx.ClientPub{
		Id:      tid,
		Topic:   topic,
		NoEcho:  true,
		Head:    nil,
		Content: bytes,
	}
	head := []byte(`\"text/x-drafty\"`)
	pub.Head["mime"] = head
	clientMsg := pbx.ClientMsg{Message: &pbx.ClientMsg_Pub{Pub: pub}}
	return &clientMsg
}
func (c *ChatBot) NoteRead(topic string, seq int) *pbx.ClientMsg {
	b := &pbx.ClientNote{
		Topic:   topic,
		What:    pbx.InfoNote_READ,
		SeqId:   int32(seq),
		Unread:  0,
		Event:   0,
		Payload: nil,
	}
	clientMsg := pbx.ClientMsg{Message: &pbx.ClientMsg_Note{Note: b}}
	return &clientMsg
}
func (c *ChatBot) OnServerDataEvent(e *ServerDataEventArgs) {
	if c.ServerDataEvent != nil {
		c.ServerDataEvent(nil, e)
	}
}
func (c *ChatBot) OnLoginEvent(isSuccess bool) {
	if isSuccess {
		if c.LoginSuccessEvent != nil {
			c.LoginSuccessEvent(c, EventArgs{})
		}
	} else {
		if c.LoginFailedEvent != nil {
			c.LoginFailedEvent(c, EventArgs{})
		}
	}
}
func (c *ChatBot) OnCtrlMessageEvent(ft FutureTypes, id string, code int, text string, topic string, parameters map[string][]byte) {
	e := CtrlMessageEventArgs{Type: ft, Id: id, Code: code, Text: text, Topic: topic, Params: parameters}
	handler := c.CtrlMessageEvent
	if handler != nil {
		handler(c, e)
	}
}
func (c *ChatBot) GetSubs(topic string, getAll bool) *pbx.ClientMsg {
	if topic == "" {
		topic = "me"
	}
	tid := c.GetNextTid()
	if getAll {
		return &pbx.ClientMsg{
			Message: &pbx.ClientMsg_Get{Get: &pbx.ClientGet{
				Id:    tid,
				Topic: topic,
				Query: &pbx.GetQuery{
					What: "sub",
					Desc: nil,
					Sub:  nil,
					Data: nil,
				},
			}},
			Extra: nil,
		}
	} else {
		return &pbx.ClientMsg{
			Message: &pbx.ClientMsg_Get{Get: &pbx.ClientGet{
				Id:    tid,
				Topic: topic,
				Query: nil,
			}},
			Extra: nil,
		}
	}
}
func (c *ChatBot) ExecFuture(tid string, code int, text string, topic string, parameters map[string][]byte) {
	if _, ok := c.OnCompletion[tid]; ok {
		bundle := c.OnCompletion[tid]
		delete(c.OnCompletion, tid)

		if code >= 200 && code < 400 {
			arg := bundle.Arg
			bundle.Action(arg, parameters)
			if bundle.Type == 4 {
				c.ClientPost(c.GetSubs("me", false))
			} else if bundle.Type == 3 {
				c.OnLoginEvent(true)
			}
			c.Log("Exec Future", fmt.Sprintf("Tid=%v Code=%v Topic=%v Text=%v Params=%v ...OK", tid, code, topic, text, parameters))
		} else {
			if bundle.Type == 3 {
				c.OnLoginEvent(false)
			}
			c.Log("Exec Future", fmt.Sprintf("Tid=%v Code=%v Topic=%v Text=%v Params=%v ...Failed", tid, code, topic, text, parameters))
		}
		c.OnCtrlMessageEvent(bundle.Type, tid, code, text, topic, parameters)
	}
}

//ChatBotPlugin
type ChatBotPlugin struct {
	//pbx.PluginServer
	//pbx.UnimplementedPluginServer
	pbx.PluginServer
}

func (c *ChatBotPlugin) Account(ctx context.Context, request *pbx.AccountEvent) (*pbx.Unused, error) {
	action := ""
	if request.Action == pbx.Crud_CREATE {
		action = "created"
	} else if request.Action == pbx.Crud_UPDATE {
		action = "updated"
	} else if request.Action == pbx.Crud_DELETE {
		action = "deleted"
	} else {
		action = "unknown"
	}
	fmt.Println(action)
	return &pbx.Unused{}, nil
}
func (c *ChatBot) InitServer() *grpc.Server {
	server := grpc.NewServer()
	chatBotPlugin := &ChatBotPlugin{}
	//server.RegisterService(&grpc.ServiceDesc{
	//	ServiceName: "server",
	//	HandlerType: nil,
	//	Methods:     nil,
	//	Streams:     nil,
	//	Metadata:    nil,
	//}, chatBotPlugin.PluginServer)
	pbx.RegisterPluginServer(server, chatBotPlugin.PluginServer)

	listenHost := strings.Split(c.Listen, ":")[0]
	listenPort, err := strconv.Atoi(strings.Split(c.Listen, ":")[1])
	if err != nil {
		panic(err)
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", listenHost, listenPort))
	if err != nil {
		panic(err)
	}
	go func() {
		err = server.Serve(lis)
	}()
	//err = server.Serve(lis)
	if err != nil {
		panic(err)
	}
	//server.Stop()
	//pbx.RegisterPluginServer(server, pbx.PluginServer(chatBotPlugin))
	//listenHost := strings.Split(c.Listen, ":")[0]
	//listenPort, err := strconv.Atoi(strings.Split(c.Listen, ":")[1])
	//lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", listenHost, listenPort))
	//if err != nil {
	//	panic(err)
	//}
	//
	//if err := server.Serve(lis); err != nil {
	//	panic(err)
	//}
	return server
}
func (c *ChatBot) InitClient() pbx.Node_MessageLoopClient {
	// ping / 2s and timeout 2s
	options := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(
			keepalive.ClientParameters{
				Time:                2 * time.Second,
				Timeout:             2 * time.Second,
				PermitWithoutStream: true,
			},
		),
	}

	conn, err := grpc.Dial(c.ServerHost, options...)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	client := pbx.NewNodeClient(conn)
	stream, err := client.MessageLoop(context.Background())
	if err != nil {
		log.Fatalf("Failed to open stream: %v", err)
	}
	c.ClientPost(c.Hello())
	c.ClientPost(c.Login(c.CookieFile, c.Schema, c.Secret))
	c.ClientPost(c.Subscribe("me"))
	return stream
}
func (c *ChatBot) Stop() {
	c.Log("Stopping", "ChatBot is exiting...Wait a second...")
	_, cancel := context.WithCancel(context.Background())
	cancel() // вызов этой функции отменит операцию
	c.Server.Stop()
	//c.Channel.

}
func (c *ChatBot) Chatbot(serverHost, listen, cookie, schema, secret string, botResponse IBotResponse) {
	rand.Seed(time.Now().UnixNano())
	c.NextTid = int64(rand.Intn(1000) + 1)
	c.ServerHost = serverHost
	c.Listen = listen
	c.CookieFile = cookie
	c.Schema = schema
	c.Secret = secret
	//c.BotResponse = botResponse
	//486
	//cancellationTokenSource = new CancellationTokenSource();
	//            sendMsgQueue = new Queue<ClientMsg>();
	c.BotResponse = botResponse
	_, cancel := context.WithCancel(context.Background())
	c.CancellationTokenSource = cancel
	c.Subscriptions = make(map[string]bool)
	c.OnCompletion = make(map[string]Future)
	c.Subscribers = make(map[string]Subscriber)

	//c.SendMessageLoop(ctx)
}
func (c *ChatBot) OnLogin(cookieFile string, parameters map[string][]byte) {
	if parameters == nil || cookieFile == "" {
		return
	}
	if user, ok := parameters["user"]; ok {
		c.BotUID = string(user)
	}

	cookieDics := make(map[string]string)
	cookieDics["schema"] = "token"
	if token, ok := parameters["token"]; ok {
		var tokenStr string
		if err := json.Unmarshal(token, &tokenStr); err != nil {
			log.Printf("OnLogin: Failed to unmarshal token: %v", err)
			return
		}
		cookieDics["secret"] = tokenStr

		var expiresStr string
		if expires, ok := parameters["expires"]; ok {
			if err := json.Unmarshal(expires, &expiresStr); err != nil {
				log.Printf("OnLogin: Failed to unmarshal expires: %v", err)
			}
		}
		cookieDics["expires"] = expiresStr
	} else {
		cookieDics["schema"] = "basic"
		if secret, ok := parameters["secret"]; ok {
			var secretStr string
			if err := json.Unmarshal(secret, &secretStr); err != nil {
				log.Printf("OnLogin: Failed to unmarshal secret: %v", err)
				return
			}
			cookieDics["secret"] = secretStr
		}
	}

	// save token for upload operation
	c.Token = cookieDics["secret"]
	cookieJson, err := json.Marshal(cookieDics)
	if err != nil {
		log.Printf("OnLogin: Failed to marshal cookie: %v", err)
		return
	}
	if err := ioutil.WriteFile(cookieFile, cookieJson, 0644); err != nil {
		log.Printf("OnLogin: Failed to write cookie file: %v", err)
	}
}
func (c *ChatBot) Upload(fileName string, redirectUrl string) (*UploadedAttachmentInfo, error) {
	if c.Token == "" {
		//not login or login failed, disable upload
		c.Log("Upload Attachment Error", "Not login, upload operation disabled...")
		return nil, nil
	}

	if fileName == "" || !FileExists(fileName) {
		c.Log("Upload Attachment Error", fmt.Sprintf("can not find file:%s", fileName))
		return nil, nil
	}

	fullFileName, _ := filepath.Abs(fileName)
	fileInfo, _ := os.Stat(fullFileName)

	attachmentInfo := UploadedAttachmentInfo{
		FullFileName: fullFileName,
		FileName:     fileInfo.Name(),
		Size:         fileInfo.Size(),
		Mime:         fmt.Sprintf("file/%s", filepath.Ext(fileName)),
	}

	restClient := resty.New().SetHostURL(c.ApiBaseUrl)
	var request *resty.Request
	if redirectUrl == "" {
		request = restClient.R()
		request.URL = c.UploadEndpoint
		request.Method = "Put"
	} else {
		request = restClient.R()
		request.URL = redirectUrl
		request.Method = "Put"
	}

	_, err := request.
		SetHeader("X-Tinode-APIKey", c.ApiKey).
		SetHeader("X-Tinode-Auth", fmt.Sprintf("Token %s", c.Token)).
		SetFile("file", fullFileName).
		Put(c.UploadEndpoint)

	if err != nil {
		//exception occurs
		c.Log("Upload Attachment Exception", fmt.Sprintf("Upload failed, Exception=%s", err.Error()))
		return nil, err
	}
	///// остановился на 1001 ChatBot.cs
	response, err := restClient.R().Execute(request.Method, request.URL)
	if response.StatusCode() >= 200 && response.StatusCode() < 400 {
		if response.StatusCode() == 200 {
			var obj map[string]interface{}
			err = json.Unmarshal([]byte(response.String()), &obj)
			if err != nil {
				panic(err)
			}
			_, ok := obj["ctrl"].(map[string]interface{})["code"].(string)
			if !ok {
				fmt.Println("217 string")
			}
			url, ok := obj["ctrl"].(map[string]interface{})["params"].(map[string]interface{})["url"].(string)
			if !ok {
				fmt.Println("221 string")
			}
			attachmentInfo.RelativeUrl = url
			c.Log("Upload Attachment Success", fmt.Sprintf("Upload file success, File=%s, RefUrl=%s", fullFileName, url))
			return &attachmentInfo, nil
		} else if response.StatusCode() == 307 {
			var tmpRedirectUrl string = ""
			//for _, h := range response.Header() {
			//	if strings.ToLower(h[]) == "location" {
			//		tmpRedirectUrl = h[0]
			//		break
			//	}
			//}
			//ChatBot.cs 1017
			log.Println("Upload Attachment Redirect", "redirect upload...")
			return c.Upload(fileName, tmpRedirectUrl)
		} else {
			c.Log("Upload Attachment Error", fmt.Sprintf("Upload failed, Http Code=%d", response.StatusCode()))
			return nil, nil
		}
	} else {
		//Failed
		c.Log("Upload Attachment Error", fmt.Sprintf("Upload failed, Http Code=%d", response.StatusCode()))
		return nil, nil
	}
}
func (c *ChatBot) ClientPost(msg *pbx.ClientMsg) {
	//newMsg := *msg
	c.SendMsgQueue = append(c.SendMsgQueue, msg)
}
func (c *ChatBot) Hello() *pbx.ClientMsg {
	tid := c.GetNextTid()
	c.AddFuture(tid, c.Future.Future(tid, Hi, func(unused string, parameters map[string][]byte) {
		c.ServerVersion(parameters)
	}, ""))

	b := &pbx.ClientMsg_Hi{Hi: &pbx.ClientHi{
		Id:         tid,
		UserAgent:  fmt.Sprintf("%s/%s %s; gRPC-csharp/%s", c.AppName, c.AppVersion, c.Platform, c.AppVersion),
		Ver:        "0.18.1",
		DeviceId:   "",
		Lang:       "EN",
		Platform:   "",
		Background: false,
	}}
	return &pbx.ClientMsg{
		Message: b,
		Extra:   nil,
	}
}
func (c *ChatBot) Login(cookieFile, scheme string, secret string) *pbx.ClientMsg {
	tid := c.GetNextTid()

	c.AddFuture(tid, c.Future.Future(tid, Login, func(fname string, parameters map[string][]byte) {
		c.OnLogin(fname, parameters)
	}, cookieFile))

	b := &pbx.ClientMsg_Login{Login: &pbx.ClientLogin{
		Id:     tid,
		Scheme: scheme,
		Secret: []byte(secret),
		Cred:   nil,
	}}
	return &pbx.ClientMsg{
		Message: b,
		Extra:   nil,
	}
}
func (c *ChatBot) ServerVersion(parameters map[string][]byte) {
	if parameters == nil {
		return
	}
	build := string(parameters["build"])
	ver := string(parameters["ver"])
	c.Log("Server Version", fmt.Sprintf("Server: %s, %s", build, ver))
}
func (c *ChatBot) Subscribe(topic string) *pbx.ClientMsg {
	tid := c.GetNextTid()
	c.AddFuture(tid, c.Future.Future(tid, Sub, func(topicName string, unused map[string][]byte) {
		c.AddSubscription(topicName)
	}, topic))

	b := &pbx.ClientMsg_Sub{Sub: &pbx.ClientSub{
		Id:       tid,
		Topic:    topic,
		SetQuery: nil,
		GetQuery: nil,
	}}

	return &pbx.ClientMsg{
		Message: b,
		Extra:   nil,
	}
}
func (c *ChatBot) AddSubscription(topic string) {
	if _, ok := c.Subscriptions[topic]; !ok {
		c.Subscriptions[topic] = true
	}
}
func (c *ChatBot) AddSubscriber(sub Subscriber) {
	log.Printf("Update Subscriber: %v", sub)
	if _, ok := c.Subscribers[sub.Topic]; ok {
		c.Subscribers[sub.Topic] = sub
	} else {
		c.Subscribers[sub.Topic] = sub
		// If it's the first subscriber for the topic, subscribe to the topic
		c.ClientPost(c.Subscribe(sub.Topic))
	}
}
func FileExists(name string) bool {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
	return true
}

//Bot Response
type BotResponse struct {
	IBotResponse
	bot ChatBot
	MsgBuilder
}

func (b *BotResponse) BotResponse(bot ChatBot) {
	b.bot = bot
}

////// остановился на Program.cs 58

func (b *BotResponse) ThinkAndReply(message *pbx.ServerData) ChatMessage {
	for _, sub := range b.bot.Subscribers {
		sub = sub
	}
	var responseMsg ChatMessage
	//msgText := message.Content
	//strings.ToValidUTF8(message.Content, msgText)
	msg := b.MsgBuilder.Parse(message)
	if msg.IsPlainText {
		if msg.Text == "image" {
			responseMsg = b.MsgBuilder.BuildImageMessage("a.png", "image/png", 47, 48, "iVBORw0KGgoAAAANSUhEUgAAAC8AAAAwCAYAAACBpyPiAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAAABmJLR0QA/wD/AP+gvaeTAAAAB3RJTUUH2wUSFzYAFqhyqAAAENNJREFUaN7FmlusncdVx39rZr69zz7HPuc4tpM4iR3bca4UkaZqi1p6L7QIVaBKwAM3lSJFBcQDT/CMhIAnHhCoEi9cXnigqqCINq2aXigladOE1tSO4zgX2/H9+Fz3d5mZtXiY2fuchPIK+2j8zf72d/nPmv9a819rLABPP/UxFpdWWFrZj9PRcQn+E843v+bFPy7ONd43OB9Kcx7npBy9w7kGJx5xDhFfmnOoes6/+iCnjr+EcxlTxSyXpopaRjWiWVHNqFo55oTmRM4RU43Z8gua499Zyv+sbnh1Z2OL6c4GH/qZLyHUz5O/Db/75G990HL+g2z5Q6p94wRx3uGDL805nHO4IKXvS/POIU5w4hBxiAgpC1//5n184H2XCN4wM8wUNcXUyKoFeNbST4Zq6eeUySmX3w1zbhy9+KfF+z/5i8/+9dc++5cFs/+dJx/ji1/4ZU6dePijKbZ/3nXb7xn6Ha+qAgYYgoEphoLVhoIlqKDMEqYR04TlREyZl1+ecOzoGmJDtfJAzqmCTmga0JzJKZVjTuisnxI5Z2KMMvRTn9JwyjS/452PP/HSH//RBy7cvrWOPPutX6Rt80PjMX819O2H1RLeeVzwhODxwc2t7L3DecF7KTPgBKlH50BEEBFASNnx1a/dw4c/+AbBKzCzvqEKqlboo9Xi2dBs5D2zkZOSUkZTJmvGSWA0mnyl6/JnFifhvP+vZ34YPvrxU7+pqf9033fiyruLxesRMxB7i9XrbMzOaUZVMc2YJnLKnH95iRPH1hCLlc8Zy3ne39sKTXKlUCbPzuf63JxJQwLsfhF36dO/8vln/J/96ftOjBv9/Rz7E6YZV0ELVh3Car+Crt/NFDGF6oCmu5Qyy+SkvHR+kZPH1xFJ82tmoEwzlst3q9ZWrYPL+qaBWc7YzA9Scjlp866fvO+boYv5KJJ/LHYD4gURwzDMXAXnEHVgAubAlaNzgJMyUAdOBHOU2RIhR0dKkRxjGbjZbLJQszJOo9KGCr7SpjqvZp3TaNZPOdEsuLe1MR8N67fbw8fu9qs5Jpx51BTMgxla3ybeAWUwVsEjgjnDeUG0AMYVxokImgt4zQlFC98pk6d1IJoNU0GtgDYtvNc54D39lNEMmjLjcV5dv90eDlub3ZLY2FtOhdcikA3DYXhMBEMxE5IWSqkTnCuOa+ZwImVGtE6MCJocOSY0RbRaXkvwKtY2Q1N1VC3NDFDB1ErLhuUMWUG14MoZsei3NrulEIcUrPKu3Osqxyt9RMgqZPOMRgcYj5bxwQOQdYc0rOF9BOcQV7wCEXJ2pBzJOeJmtKHQJquSc8BxkJFbAgc5Zfphk2G4DZLLTGWDrNUnDEulbzkThxSCmYlpPYlhYmiliJrDVBBZYN++46wun2RhvIJzAQNimrLTvkI3nIfQ4UQq5ymWTwmNsVKxctyMlBZYaE6xNDlBExYRQDXR9Rusb15ge/tVzLpqfUWVPcCtrtYmwWqYI+caIwvP1UtdmAL79x1haXyUN6623NrYZkiO8Ug4uNJw16GjqEv0/VlCiDgnIGC5gLecUNE5XVIKjJvjNO4oFy9vcWtjjX4wRkE5uOI5fOAoqe/Z2r6AkNBsMG+FOmjGTAnUGGopF4tjSHE7NAtutIL3h3nlcs/gjnLX8VNMFpcZ+ik3r73GzqVL3H3HISyvkO1qcW63x/IpYaKgVKc8SOYQF671ZH+MIyfuZzRepJ1ucv3KebYvX+Tw6mGEa+iwVtiQZ448o04GVQKUaSAbJloiuggOQfF4W+bWesAtPsJPvO2nOHjoCD40aE6s3/cg588+y8310+ybrGByHUdGVOoLq2GkrKw5etRW2NgKjJcf49Qj72L1wCGcD+QUuXXvA7x4+t+4tX6ahmVyXkMoIbQ4b5kB07LeOFQRzZU6ClmR+tKcjGEYM3Anx0+9k4OH7il8VxAJrBy4k2Mn3072R5hOF8gDWKqLStUqtneRGSjX+SMcO/l2Vg7ciUh5nnOBg4fu4fipdzJwJ8MwJqcaXSquEnVyxVtpQyqWL6to9ThPARAWWFq9n+WVQ8SYafuOlDIheBbGDUv7V1laPcb06hmyKeYz4sCSKytmyoCWKDMoKS+ydOgYS/tXGYZI18f58ybjhuWVQyyt3k979QwWM3iFTME3s34qoTOUGF41ibgS1x2ISlV4wurkDtTgey9e4W++/DLnLm1w4sh+fv2jD/D2Bw4yHq+wmR2ZiGlGTdBkdTnP6Iw2OZKyYzxeQZPyzIuX+duvvMwrV7Z46L4VfuOnH+DRYyuMJ3ewmQWXEkhd4JQa/xUzj6EEMUNUKSvIzPJgQbCsxK7D+RH9EPmHr5zlH88dZlh4hP+8cBP31Dke+tQ7EB+IcYrZAKmsF5Zc0SUpFykNWDJimiI+sNMN/P1T5/j8q0dI48f4wbmrLMhZ/vBXn8D5EbHrGOWEeSDZ3PpowStmuJmQmgmnmWCylHFkUlwnxW3iMHDpym2GtITGEUn3cenKOkPfk+IUS+uIFS0+U5ozsWZWhJZYwtI6KU4Z+p5LV9ZJug+NI4a0xKUrt4nDQIrbpLiOow4+z9SqzjMyTHFlSiqX1KraLSuKF2XETfqt14lDz6mjqyy1l3AbbzDZvsiD962QU6TfvsTYXSW4DHVhsaoMZ33UCC4zdlfpty+RU+TB+1aYbF/EbbzBUnuJU0dXiUNPv/U6I27iZSboCi5mUUfLuYDtfinawxAtC41zMGaduP5duu2H+YX3P4D3r/DaxR9y7z2rfOK9J0n9LdLmd5i46zgpz0KoCYdWjV8WKScwcddJm98hHTjJL334JJPxBS6/cYX7j97BJ977AN32FeL6dxmzXldemytRNZgZG7MS58W08MgJYlUluqLTGhmw7efYuX4Hh4/+HJ/5+YfxfoTmyNDdYOvKl/Dtt2lci6jDpMriKrjIBs7m+UzjWmi/TXdzlXuPfIzf++RjON+Q80C/8wa3L/4rsv1ceW8VcqIgZqVpwQtGMGajsrm/ijNEQQW8CE43iDeeYpsbcPBRRot3ELsNuo1z5NvP4fV2tbbO9fw8MujuzM7AeLtJXvsCrV3BVh6iWVhhmK6xfesM8cbzON3EsLfo/j1qtIq8IBhObJ45zeWBzJUOYhkbbtNd/yZ5/Tv4EGp21OMslhmrqhgBEcOFIt5dUJzTOXC1YkXLN4m3vkq6/S3MAjkl4tBhMRZ5VbHMcOzFN8MbSgyrkg9mt8xTVqmyGIOcBtCBG9twa1Pwvibd8ySk3I+DlOHW+sCLr0LwUrVyzaaqNDYbyHng4LKxuq8soGZSdX2x+mwgsCefrliLw87+9v5mVjI/CnCp572HZ8+O+Nw3lhg1brfw8yN6RuTM+ZXZa/d8bP7vEJVPvn+Hn333QMrspotvwiTzysPsDzOCYmS1MmqpgKVwviY2iBolZRWyGk0z4uDBQyyMm13A8uYB/O8f251koOsjTRPJ2pWVuea0JQOd9XfLJaqQ1VCMEHWRnbiPjT7jQ6m/OF+p4AuPxe2GztFgmFtmZWWZ8SjMyPk/YcueM3soWWzK/L7xkDHXsr7jGLJQBeM8LazrUanpKGgyfPRE3SYcPvkkj37k4/T9gADNaIT3rpTTfEn3mlGDc+WcE+EJdXzKQonDFZhI4b/V72aGq87wpnP1OTrLh9UYN+Ck1mxqwyBrKXfknOdUjikxHo+4MHyRcP3GJmfOvc7mxgaaEpOlJUIILC0u4kPx58lkgdA0jEbjeXVApFTNxAk5ZURmPuERgRgT3jtUjRA8ZpBSwnmHmeG9r3UcnQ9azej7jhQjwxBLJS0ltnd2yDnTty04x/LKCtdvbOIA+q6j6ztG4xFt22JmdH1HzhnvPX0/oGr0fV/LeUZKCVUlDgPOuVrZKho+xoj3nhgTZkaMkZwTzjniEIsFh2HOppRKntv3fX2WklIipUTXdTRNoO86sipd39F3XYk2ZobzHifFSqPRCOc93vlaHYPQBJxItZrinJtbuFhN59XhXZpbrWe6SpvSyn2uzlRZ1LwvM+O8Q3N5fhMCMUWcd6SshKZQV81qCV1x9Zm0bVu4BvR9h2H0XYdJyT1jSqgZwzCUEnZK84JQignnhBgjYLUSnCt9hhI0Y5yHu5Rm1o8IQt5Tk4wx4pyjHwZMjWGIxDjgnaOdTjEzhn4os4SUaWuahhRL9r8wHpNiJDQj4hAJIRC8J+dMCIGUEk0TijNiiAg5K03TVIlQHdegCc2cfjnlWqwKZM344Ek5EULAOSnWdMUw4/EYNaMGOrIqi0uLpJTwzrGwsICzGrebpsEql7uuL/RxMrdayolRsxvXC58hhDCvKsc4zClVtIkyxAHvfYkyQrVyqvfJ3JFVbff51XFFYDQaVSWpTHd2aJqGZjTC1Irl27ala1tCaJi2Ld47ptOWnDOIkFJEgLbtMCs0moW0rjpxoYwQY2IYIiKuUswxDJGU0vx3ELq+nwu2lEuFrG278qwUqyNnptOWEALTtsU5P8cqIoRhiHRdz42bt9jZmeKbQNt2TBYmbGxuMVlcxDlhtLVNCA3rss5o1JBSro5WFpNZxAnBV767GscVN6OMd/OYX6hWrp/F8ZQzwzBgpnRtV4zTdcRUBt9Pu7KyhhE705Zw8fzzlqeX7cLFyywuTopHq7J//366tmP/8v7C2eAZj8fEOLC4uEjfdWX6aoo3Go3pupaFhQlDdbohjnGyw2SyQNu2jEajus+kNM2Irm1ZmEzmIdIM2q5j3IzY2tqiaRr6vidWnm9ubjJZXOLA0sQuX7pmod94Jts+P6x4JtoViwSBfq04y7Qv9XcR6CgSoVuzuoE2q81DTznubM+sKrxwdpnHH91EtxVMmFYnNqCt+qVdq+K3FoIB+tkxl2tD1Tz7RJDOGAUGFzWH8aJrH36o2XTCJA67YMRRdU5VuVXbzM7LbB9qz/VSr0EgJuHGxhLve3dLE6zqlfluUEnpKuC5AKvf56XwzO75en0zAjU2r6wNXTh9Lt96/BF7/eQxd9e0K9aeg/GCOKPKkfnGmZMyqPlAKmgnNYlxEAchBNg3EZoRc22utqsa58CylNy/KseifcBq2XAGXg2WFoSXL+rrp1/KN90rl/Pl759Jz/eddU1Td2/qJiZW96UqoLL5UHJSqbmquD1NDOcNcQZ+nvMVA3hDxN5yfRWfzubPFrdHpc5q+hVT00DXWf+DM+mFVy7nyw64/vXnhmefP53OkTDva5WhDmA2XbMkZddqRVPPkwSdJQ+1v4ci8xyWvflsuX8mfXXPO+Y0qTiUkgSRsOdPpxef/u7wDHDde8fQ9ujFa7p/MuKue1fcfj/C4d6SFtU9NSnbU1DTv7kcdrvXISWl+/7ZBX780Rbv96R+Mttk2DWO1gQl700RZ2lAKZtiO6T/+EF67Z++Hr+4tmFf8o7XvRkqMN1prT/7qvprazq6e9ktLgaaRkRGDeKBIBDqsQG8wGjveQpTPKVJFk6/uMgTD7U0znAKTss1TgtTvBa2hNn3PS0I2ICl1vTaVdv63NPDi19+Jn1tY9v+ReD7arSyx74HgMeB9x++w73j8Qf9iSMHZHl5QSYy32+w+dZN3fzbTZpkzywIpCz8+5l9vOfR7fp/D/YkVHv6yh5Kms0dzgTb6mmv3NbN58/lV26s6feAbwDPA7erK+wmbsAicH8dxKPA3cD+asz/608GtoCrwBngBeA1YEqNJz8qY/bAKnAXcLAO6P8L/BS4BVwD1uu5+ee/Aa0AqnTFGSHgAAAAJXRFWHRkYXRlOmNyZWF0ZQAyMDE4LTA2LTI4VDIyOjM5OjU1KzA4OjAwwmBIVQAAACV0RVh0ZGF0ZTptb2RpZnkAMjAxMS0wNS0xOFQyMzo1NDowMCswODowMDef1rEAAABDdEVYdHNvZnR3YXJlAC91c3IvbG9jYWwvaW1hZ2VtYWdpY2svc2hhcmUvZG9jL0ltYWdlTWFnaWNrLTcvL2luZGV4Lmh0bWy9tXkKAAAAGHRFWHRUaHVtYjo6RG9jdW1lbnQ6OlBhZ2VzADGn/7svAAAAF3RFWHRUaHVtYjo6SW1hZ2U6OkhlaWdodAA2MLuNbZ0AAAAWdEVYdFRodW1iOjpJbWFnZTo6V2lkdGgANTkR00Z3AAAAGXRFWHRUaHVtYjo6TWltZXR5cGUAaW1hZ2UvcG5nP7JWTgAAABd0RVh0VGh1bWI6Ok1UaW1lADEzMDU3MzQwNDCrCw2wAAAAEXRFWHRUaHVtYjo6U2l6ZQA1OTQzQjpo3pUAAABgdEVYdFRodW1iOjpVUkkAZmlsZTovLy9ob21lL3d3d3Jvb3QvbmV3c2l0ZS93d3cuZWFzeWljb24ubmV0L2Nkbi1pbWcuZWFzeWljb24uY24vc3JjLzUwOTIvNTA5Mjg0LnBuZ91LspIAAAAASUVORK5CYII=", "this is a image by chatbot")
		} else if msg.Text == "file" {
			responseMsg = b.MsgBuilder.BuildFileMessage("a.txt", "text/plain", "ZHNmZHMKZmRzZmRzZnNkZgpm5Y+N5YCS5piv56a75byA5oi/6Ze055qE5LiK6K++5LqGCg==", "this is a file by chatbot")
		} else if msg.Text == "more" {
			builder := b.MsgBuilder.MsgBuilder()
			builder.AppendText("Hi,this is bold\n", true, false, false, false, false, false, false, false, false, "", "", "")
			builder.AppendText("Hi,this is italic\n", false, true, false, false, false, false, false, false, false, "", "", "")
			builder.AppendText("Hi,this is deleted\n", false, false, true, false, false, false, false, false, false, "", "", "")
			builder.AppendText("int a=100;\nint b=a*100-90;\n", false, false, false, true, false, false, false, false, false, "", "", "")
			builder.AppendText("https://google.com\n", false, false, false, false, true, false, false, false, false, "", "", "")
			builder.AppendText("baidu.com\n", false, false, false, false, true, false, false, false, false, "", "", "")
			builder.AppendText("@tinode\n", false, false, false, false, false, true, false, false, false, "", "", "")
			builder.AppendText("#Tinode\n", false, false, false, false, false, false, true, false, false, "", "", "")
			builder.AppendText("\n\nnext is image\n", false, false, false, false, false, false, false, false, false, "", "", "")
			builder.AppendImage("a.png", "image/png", 0, 0, "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ACV0RVh0ZGF0ZTptb2RpZnkAMjAxMS0wNS0xOFQyMzo1NDowMCswODowMDef1rEAAABDdEVYdHNvZnR3YXJlAC91c3IvbG9jYWwvaW1hZ2VtYWdpY2svc2hhcmUvZG9jL0ltYWdlTWFnaWNrLTcvL2luZGV4Lmh0bWy9tXkKAAAAGHRFWHRUaHVtYjo6RG9jdW1lbnQ6OlBhZ2VzADGn/7svAAAAF3RFWHRUaHVtYjo6SW1hZ2U6OkhlaWdodAA2MLuNbZ0AAAAWdEVYdFRodW1iOjpJbWFnZTo6V2lkdGgANTkR00Z3AAAAGXRFWHRUaHVtYjo6TWltZXR5cGUAaW1hZ2UvcG5nP7JWTgAAABd0RVh0VGh1bWI6Ok1UaW1lADEzMDU3MzQwNDCrCw2wAAAAEXRFWHRUaHVtYjo6U2l6ZQA1OTQzQjpo3pUAAABgdEVYdFRodW1iOjpVUkkAZmlsZTovLy9ob21lL3d3d3Jvb3QvbmV3c2l0ZS93d3cuZWFzeWljb24ubmV0L2Nkbi1pbWcuZWFzeWljb24uY24vc3JjLzUwOTIvNTA5Mjg0LnBuZ91LspIAAAAASUVORK5CYII=")
			builder.AppendText("\n\nnext is file\n", false, false, false, false, false, false, false, false, false, "", "", "")
			builder.AppendFile("a.txt", "text/plain", "ZHNmZHMKZmRzZmRzZnNkZgpm5Y+N5YCS5piv56a75byA5oi/6Ze055qE5LiK6K++5LqGCg==")
			responseMsg = builder.Message
		} else if msg.Text == "attach" {
			uploadInfo, err := b.bot.Upload("./libgrpc_csharp_ext.x64.so", "")
			if err != nil {
				panic(err)
			}
			if uploadInfo != nil {
				responseMsg = b.MsgBuilder.BuildAttachmentMessage(*uploadInfo, "This is a larget attachment file")
			} else {
				responseMsg = b.MsgBuilder.BuildTextMessage("I try to send you a larget attach file, but I am sorry I failed...")
			}
		} else if msg.Text == "form" {
			// дальще дописать
		}
	} else {
		responseMsg = msg
		mentions := msg.GetMentions()
		for _, m := range mentions {
			fmt.Printf("Mentions:%v\n", m.Val)
		}
		images := msg.GetImages()
		for _, image := range images {
			fmt.Printf("Image:Name=%s Mime=%s\n", image.Name, image.Mime)
		}
		hashTags := msg.GetHashTags()
		for _, hash := range hashTags {
			fmt.Println("HashTags:", hash.Val)
		}
		links := msg.GetLinks()
		for _, link := range links {
			fmt.Println("Links:", link.Url)
		}
		files := msg.GetGenericAttachment()
		for _, f := range files {
			fmt.Printf("Image: Name=%s Mime=%s", f.Name, f.Mime)
		}
	}
	return responseMsg
}

//IBotResponse

type IBotResponse interface {
	ThinkAndReply(message *pbx.ServerData) (*ChatMessage, error)
}

// ChatMessage

type ChatMessage struct {
	MessageBase
	Text        string       `json:"txt"`
	Fmt         []FmtMessage `json:"fmt"`
	Ent         []EntMessage `json:"ent"`
	IsPlainText bool         `json:"-"`
	MessageType string       `json:"-"`
}

func (c *ChatMessage) ChatMessage() {
	c.Text = ""
	c.MessageType = "unknown"
}
func (cm *ChatMessage) GetFormattedText() string {
	if cm.Text == "" {
		return ""
	}

	textArray := []rune(cm.Text)

	if cm.Fmt != nil {
		for _, fmtv := range cm.Fmt {
			if fmtv.Tp == "BR" && fmtv.At != nil {
				textArray[*fmtv.At] = '\n'
			}
		}
	}

	return string(textArray)
}
func (c ChatMessage) GetEntDatas(tp string) []EntData {
	ret := make([]EntData, 0)
	if c.Ent != nil {
		for _, ent := range c.Ent {
			if ent.Tp == tp {
				ret = append(ret, ent.Data)
			}
		}
	}
	return ret
}

// Get mentioned users

func (c *ChatMessage) GetMentions() []EntData {
	return c.GetEntDatas("MN")
}

// Get images

func (c *ChatMessage) GetImages() []EntData {
	return c.GetEntDatas("IM")
}

// Get hashtags

func (c *ChatMessage) GetHashTags() []EntData {
	return c.GetEntDatas("HT")
}

// Get links

func (c *ChatMessage) GetLinks() []EntData {
	return c.GetEntDatas("LN")
}

// Get generic attachment

func (c *ChatMessage) GetGenericAttachment() []EntData {
	return c.GetEntDatas("EX")
}

//EntMessage

type EntMessage struct {
	MessageBase
	Tp   string  `json:"tp"`
	Data EntData `json:"data"`
}

// EntData

type EntData struct {
	MessageBase
	Mime   string      `json:"mime"`
	Val    interface{} `json:"val"`
	Url    string      `json:"url"`
	Ref    string      `json:"ref"`
	Width  *int        `json:"width"`
	Height *int        `json:"height"`
	Name   string      `json:"name"`
	Size   *int        `json:"size"`
	Act    string      `json:"act"`
}

// fmtMessage
type FmtMessage struct {
	MessageBase
	At  *int   `json:"at"`
	Len *int   `json:"len"`
	Tp  string `json:"tp"`
	Key *int   `json:"key"`
}

// future

type Future struct {
	Tid    string
	Arg    string
	Type   FutureTypes
	Action func(string, map[string][]byte)
}

func (f *Future) Future(tid string, typ FutureTypes, action func(string, map[string][]byte), arg string) Future {
	return Future{Tid: tid, Type: typ, Action: action, Arg: arg}
}

// Subscriber

type Subscriber struct {
	MessageBase
	UserId    string
	Topic     string
	UserName  string
	PhotoData string
	Type      string
	Online    bool
	PhotoType string
}

func NewSubscriber(userId, topic, username, typeStr, photoData, photoType string, online bool) *Subscriber {
	mb := MessageBase{}

	return &Subscriber{
		MessageBase: mb,
		UserId:      userId,
		Topic:       topic,
		UserName:    username,
		PhotoData:   photoData,
		Type:        typeStr,
		Online:      online,
		PhotoType:   photoType,
	}
}

// MsgBuilder
type MsgBuilder struct {
	Message ChatMessage
}

func (ms *MsgBuilder) MsgBuilder() *MsgBuilder {
	return &MsgBuilder{
		Message: ChatMessage{
			Ent: []EntMessage{},
			Fmt: []FmtMessage{},
		},
	}
}

func (ms *MsgBuilder) ReSet() {
	ms.Message = ChatMessage{}
	ms.Message.Ent = []EntMessage{}
	ms.Message.Fmt = []FmtMessage{}
}

func (ms *MsgBuilder) AppendText(text string, isBold bool, isItalic bool, isDeleted bool, isCode bool, isLink bool, isMention bool, isHashTag bool, isForm bool, isButton bool, buttonDataName string, buttonDataAct string, buttonDataVal string) {
	baseLen := len(ms.Message.Text)
	ms.Message.Text += text
	if strings.Contains(text, "\n") {
		for i := 0; i < len(text); i++ {
			if text[i] == '\n' {
				newAt := baseLen + i
				newLen := 1
				fmtv := FmtMessage{At: &newAt, Tp: "BR", Len: &newLen}
				ms.Message.Fmt = append(ms.Message.Fmt, fmtv)
			}
		}
	}
	leftLen := baseLen + len(strings.TrimLeft(text, " "))
	subLen := len(text) - len(strings.TrimRight(text, " "))
	validLen := len(ms.Message.Text) - leftLen - subLen

	if isBold {
		fmtv := FmtMessage{Tp: "ST", At: &leftLen, Len: &validLen}
		ms.Message.Fmt = append(ms.Message.Fmt, fmtv)
	}

	if isItalic {
		fmtv := FmtMessage{Tp: "EM", At: &leftLen, Len: &validLen}
		ms.Message.Fmt = append(ms.Message.Fmt, fmtv)
	}

	if isDeleted {
		fmtv := FmtMessage{Tp: "DL", At: &leftLen, Len: &validLen}
		ms.Message.Fmt = append(ms.Message.Fmt, fmtv)
	}

	if isCode {
		fmtv := FmtMessage{Tp: "CO", At: &leftLen, Len: &validLen}
		ms.Message.Fmt = append(ms.Message.Fmt, fmtv)
	}

	if isLink {
		lenNew := len(ms.Message.Ent)
		fmtv := FmtMessage{At: &leftLen, Len: &validLen, Key: &lenNew}
		ms.Message.Fmt = append(ms.Message.Fmt, fmtv)
		url := strings.ToLower(strings.TrimSpace(text))
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "http://" + strings.TrimSpace(text)
		}
		ent := EntMessage{Tp: "LN", Data: EntData{Url: url}}
		ms.Message.Ent = append(ms.Message.Ent, ent)
	}

	if isMention {
		lenNew := len(ms.Message.Ent)
		fmtv := FmtMessage{At: &leftLen, Len: &validLen, Key: &lenNew}
		ms.Message.Fmt = append(ms.Message.Fmt, fmtv)
		mentionName := text[1:]
		ent := EntMessage{Tp: "MN", Data: EntData{Val: mentionName}}
		ms.Message.Ent = append(ms.Message.Ent, ent)
	}

	if isHashTag {
		lenNew := len(ms.Message.Ent)
		fmtv := FmtMessage{At: &leftLen, Len: &validLen, Key: &lenNew}
		ms.Message.Fmt = append(ms.Message.Fmt, fmtv)
		hashTag := strings.TrimSpace(text)
		ent := EntMessage{Tp: "HT", Data: EntData{Val: hashTag}}
		ms.Message.Ent = append(ms.Message.Ent, ent)
	}

	if isForm {
		fmtv := FmtMessage{At: &leftLen, Len: &validLen, Tp: "FM"}
		ms.Message.Fmt = append(ms.Message.Fmt, fmtv)
	}

	if isButton {
		btnName := buttonDataName
		if btnName == "" {
			buttonDataName = strings.ToLower(strings.TrimSpace(text))
		}
		lenNew := len(ms.Message.Ent)
		fmtv := FmtMessage{At: &leftLen, Len: &validLen, Tp: "FM", Key: &lenNew}
		ms.Message.Fmt = append(ms.Message.Fmt, fmtv)
		//btnText := strings.TrimSpace(text)
		ent := EntMessage{Tp: "BN", Data: EntData{Name: buttonDataName, Act: buttonDataAct, Val: buttonDataVal}}
		ms.Message.Ent = append(ms.Message.Ent, ent)
	}
}
func (ms *MsgBuilder) AppendTextLine(text string, isBold, isItalic, isDeleted, isCode, isLink, isMention, isHashTag, isForm, isButton bool, buttonDataName, buttonDataAct, buttonDataVal string) {
	ms.AppendText(text+"\n", isBold, isItalic, isDeleted, isCode, isLink, isMention, isHashTag, isForm, isButton, buttonDataName, buttonDataAct, buttonDataVal)
}
func (ms *MsgBuilder) AppendImage(imageName string, mime string, width int, height int, imageBase64 string) {
	AtN := len(ms.Message.Text)
	NLen := 1
	KeyN := len(ms.Message.Ent)
	ms.Message.Fmt = append(ms.Message.Fmt, FmtMessage{
		At:  &AtN,
		Len: &NLen,
		Key: &KeyN,
	})

	ms.Message.Ent = append(ms.Message.Ent, EntMessage{
		Tp: "IM",
		Data: EntData{
			Mime:   mime,
			Width:  &width,
			Height: &height,
			Name:   imageName,
			Val:    imageBase64,
		},
	})
}
func (ms *MsgBuilder) AppendFile(fileName string, mime string, contentBase64 string) {
	AtN := len(ms.Message.Text)
	LenN := 0
	KeyN := len(ms.Message.Ent)
	ms.Message.Fmt = append(ms.Message.Fmt, FmtMessage{
		At:  &AtN,
		Len: &LenN,
		Key: &KeyN,
	})

	ms.Message.Ent = append(ms.Message.Ent, EntMessage{
		Tp: "EX",
		Data: EntData{
			Mime: mime,
			Name: fileName,
			Val:  contentBase64,
		},
	})
}
func (ms *MsgBuilder) AppendAttachment(attachmentInfo UploadedAttachmentInfo) {
	AtN := len(ms.Message.Text)
	LenN := 1
	KeyN := len(ms.Message.Ent)
	ms.Message.Fmt = append(ms.Message.Fmt, FmtMessage{
		At:  &AtN,
		Len: &LenN,
		Key: &KeyN,
	})
	SizeN := int(attachmentInfo.Size)
	ms.Message.Ent = append(ms.Message.Ent, EntMessage{
		Tp: "EX",
		Data: EntData{
			Mime: attachmentInfo.Mime,
			Name: attachmentInfo.FileName,
			Ref:  attachmentInfo.RelativeUrl,
			Size: &SizeN,
		},
	})

}
func (ms *MsgBuilder) Parse(message *pbx.ServerData) ChatMessage {
	var chatMsg ChatMessage

	if _, ok := message.Head["mime"]; ok {
		err := json.Unmarshal([]byte((message.Content)), &chatMsg)
		if err != nil {
			panic(err)
		}
		chatMsg.IsPlainText = false
	} else {
		chatMsg = ChatMessage{Text: string(message.Content)}
		chatMsg.IsPlainText = true
	}

	if strings.HasPrefix(message.Topic, "usr") {
		chatMsg.MessageType = "user"
	} else if strings.HasPrefix(message.Topic, "grp") {
		chatMsg.MessageType = "group"
	}

	return chatMsg
}
func (ms *MsgBuilder) TryParse(message *pbx.ServerData) (ChatMessage, bool) {
	var chatMsg ChatMessage
	chatMsg = ms.Parse(message)
	if reflect.DeepEqual(chatMsg, ChatMessage{}) {
		return ChatMessage{}, false
	}
	return chatMsg, true
}
func (ms *MsgBuilder) BuildTextMessage(text string) ChatMessage {
	msg := ChatMessage{Text: text, Ent: []EntMessage{}, Fmt: []FmtMessage{}}
	if strings.Contains(text, "\n") {
		NLen := 1
		for i, c := range text {
			if c == '\n' {
				fmtv := FmtMessage{At: &i, Tp: "BR", Len: &NLen}
				msg.Fmt = append(msg.Fmt, fmtv)
			}
		}
	}
	return msg
}
func (ms *MsgBuilder) BuildImageMessage(imageName, mime string, width, height int, imageBase64 string, text string) ChatMessage {
	msg := ChatMessage{Text: text, Ent: []EntMessage{}, Fmt: []FmtMessage{}}
	entData := EntData{Mime: mime, Width: &width, Height: &height, Name: imageName, Val: imageBase64}
	ent := EntMessage{Tp: "IM", Data: entData}
	msg.Ent = append(msg.Ent, ent)
	AtN := len(text)
	NLen := 1
	Nkey := 0
	fmtv := FmtMessage{At: &AtN, Len: &NLen, Key: &Nkey}
	msg.Fmt = append(msg.Fmt, fmtv)

	if strings.Contains(text, "\n") {
		for i := 0; i < len(text); i++ {
			if text[i] == '\n' {
				fmtc := FmtMessage{At: &i, Tp: "BR", Len: &NLen}
				msg.Fmt = append(msg.Fmt, fmtc)
			}
		}
	}

	return msg
}
func (ms *MsgBuilder) BuildFileMessage(fileName string, mime string, contentBase64 string, text string) ChatMessage {
	msg := ChatMessage{Text: text, Ent: []EntMessage{}, Fmt: []FmtMessage{}}
	msg.Ent = append(msg.Ent, EntMessage{
		Tp: "EX",
		Data: EntData{
			Mime: mime,
			Name: fileName,
			Val:  contentBase64,
		},
	})
	AtN := len(text)
	LenKey := 0
	msg.Fmt = append(msg.Fmt, FmtMessage{
		At:  &AtN,
		Len: &LenKey,
		Key: &LenKey,
	})
	LenN := 1
	if strings.Contains(text, "\n") {
		for i := 0; i < len(text); i++ {
			if text[i] == '\n' {
				fmtMsg := FmtMessage{At: &i, Tp: "BR", Len: &LenN}
				msg.Fmt = append(msg.Fmt, fmtMsg)
			}
		}
	}
	return msg
}
func (ms *MsgBuilder) BuildAttachmentMessage(attachmentInfo UploadedAttachmentInfo, text string) ChatMessage {
	SizeN := int(attachmentInfo.Size)
	AtN := len(text)
	LenN := 1
	KeyN := 0
	msg := ChatMessage{
		Text: text,
		Ent: []EntMessage{
			{
				Tp: "EX",
				Data: EntData{
					Mime: attachmentInfo.Mime,
					Name: attachmentInfo.FileName,
					Ref:  attachmentInfo.RelativeUrl,
					Size: &SizeN,
				},
			},
		},
		Fmt: []FmtMessage{
			{
				At:  &AtN,
				Len: &LenN,
				Key: &KeyN,
			},
		},
	}

	if strings.Contains(text, "\n") {
		for i := 0; i < len(text); i++ {
			if text[i] == '\n' {
				fmtv := FmtMessage{
					At:  &i,
					Tp:  "BR",
					Len: &LenN,
				}
				msg.Fmt = append(msg.Fmt, fmtv)
			}
		}
	}

	return msg
}

// MessageBase

type MessageBase struct{}

func (m *MessageBase) ToString() string {
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(b)
}

// ServerDataEventArgs

type ServerDataEventArgs struct {
	Data *pbx.ServerData
}

func (s *ServerDataEventArgs) ServerDataEventArgs(data *pbx.ServerData) *ServerDataEventArgs {
	return &ServerDataEventArgs{Data: data}
}

// CtrlMessageEventArgs

type CtrlMessageEventArgs struct {
	EventArgs
	Type  FutureTypes
	Id    string
	Code  int
	Text  string
	Topic string
	//HasError  bool
	Params map[string][]byte
}

type FutureTypes int

const (
	Unknown FutureTypes = 1
	Hi      FutureTypes = 2
	Login   FutureTypes = 3
	Sub     FutureTypes = 4
	Get     FutureTypes = 5
	Pub     FutureTypes = 6
	Note    FutureTypes = 7
	Leave   FutureTypes = 8
)

func (c *CtrlMessageEventArgs) HasError() bool {
	return !(c.Code >= 200 && c.Code < 400)
}
func NewCtrlMessageEventArgs(t FutureTypes, id string, code int, text string, topic string, params map[string][]byte) *CtrlMessageEventArgs {
	return &CtrlMessageEventArgs{
		Type:   t,
		Id:     id,
		Code:   code,
		Text:   text,
		Topic:  topic,
		Params: params,
	}
}

//ServerMetaEventArgs

type ServerMetaEventArgs struct {
	Meta *pbx.ServerMeta
}

func (s *ServerMetaEventArgs) ServerMetaEventArgs(meta *pbx.ServerMeta) *ServerMetaEventArgs {
	return &ServerMetaEventArgs{
		Meta: meta,
	}
}

// EventArgs

type EventArgs struct{}

type ServerPresEventArgs struct {
	Pres *pbx.ServerPres
}

func (s *ServerPresEventArgs) ServerPresEventArgs(pres *pbx.ServerPres) *ServerPresEventArgs {
	s.Pres = pres
	return s
}

type ServerPres struct {
	FileName string
	Mime     string
	Size     int64
}

//UploadedAttach

type UploadedAttachmentInfo struct {
	FullFileName string
	FileName     string
	Mime         string
	Size         int64
	RelativeUrl  string
}

// Геттеры

func (c *ChatBot) GetAppName() string {
	c.AppName = "ChatBot"
	return c.AppName
}
func (c *ChatBot) GetAppVersion() string {
	c.AppVersion = "0.18.1"
	return c.AppVersion
}
func (c *ChatBot) GetLibVersion() string {
	c.LibVersion = "0.18.1"
	return c.LibVersion
}
func (c *ChatBot) GetPlatform() string {
	c.Platform = "web"
	return c.Platform
}
func (c *ChatBot) GetUploadEndpoint() string {
	c.UploadEndpoint = "/v0/file/u"
	return c.UploadEndpoint
}
func (c *ChatBot) GetDownloadEndpoint() string {
	c.DownloadEndpoint = "/v0/file/s"
	return c.DownloadEndpoint
}

func (f *Future) GetTid() string {
	return f.Tid
}

func (f *Future) GetArg() string {
	return f.Arg
}

func (f *Future) GetType() FutureTypes {
	return f.Type
}

func (f *Future) GetAction() func(string, map[string][]byte) {
	return f.Action
}
func (s *Subscriber) GetUserID() string {
	return s.UserId
}

func (s *Subscriber) SetUserID(userID string) {
	s.UserId = userID
}

func (s *Subscriber) GetTopic() string {
	return s.Topic
}

func (s *Subscriber) SetTopic(topic string) {
	s.Topic = topic
}

func (s *Subscriber) GetUserName() string {
	return s.UserName
}

func (s *Subscriber) SetUserName(userName string) {
	s.UserName = userName
}

func (s *Subscriber) GetPhotoData() string {
	return s.PhotoData
}

func (s *ServerDataEventArgs) GetData() *pbx.ServerData {
	return s.Data
}

func (t *CtrlMessageEventArgs) GetParams() map[string][]byte {
	return t.Params
}
func (t *CtrlMessageEventArgs) GetId() string {
	return t.Id
}

func (t *CtrlMessageEventArgs) GetCode() int {
	return t.Code
}

func (t *CtrlMessageEventArgs) GetText() string {
	return t.Text
}

func (t *CtrlMessageEventArgs) GetTopic() string {
	return t.Topic
}

func (t *CtrlMessageEventArgs) GetType() FutureTypes {
	return t.Type
}

func (mc *UploadedAttachmentInfo) GetFileName() string {
	return mc.FileName
}

func (mc *UploadedAttachmentInfo) GetMime() string {
	return mc.Mime
}

func (mc *UploadedAttachmentInfo) GetSize() int64 {
	return mc.Size
}

func (mc *UploadedAttachmentInfo) GetRelativeUrl() string {
	return mc.RelativeUrl
}

// Сеттеры

func (f *Future) SetTid(tid string) {
	f.Tid = tid
}

func (f *Future) SetArg(arg string) {
	f.Arg = arg
}

func (f *Future) SetType(typ FutureTypes) {
	f.Type = typ
}

func (f *Future) SetAction(action func(string, map[string][]byte)) {
	f.Action = action
}
func (s *Subscriber) SetPhotoData(photoData string) {
	s.PhotoData = photoData
}

func (s *Subscriber) GetType() string {
	return s.Type
}

func (s *Subscriber) SetType(subType string) {
	s.Type = subType
}

func (s *Subscriber) IsOnline() bool {
	return s.Online
}

func (s *Subscriber) SetOnline(online bool) {
	s.Online = online
}

func (s *Subscriber) GetPhotoType() string {
	return s.PhotoType
}

func (s *Subscriber) SetPhotoType(photoType string) {
	s.PhotoType = photoType
}

func (s *ServerDataEventArgs) SetData(data *pbx.ServerData) {
	s.Data = data
}

func (t *CtrlMessageEventArgs) SetParams(params map[string][]byte) {
	t.Params = params
}
func (t *CtrlMessageEventArgs) SetId(id string) {
	t.Id = id
}

func (t *CtrlMessageEventArgs) SetCode(code int) {
	t.Code = code
}

func (t *CtrlMessageEventArgs) SetText(text string) {
	t.Text = text
}

func (t *CtrlMessageEventArgs) SetTopic(topic string) {
	t.Topic = topic
}

func (t *CtrlMessageEventArgs) SetType(typE FutureTypes) {
	t.Type = typE
}

func (mc *UploadedAttachmentInfo) SetFileName(name string) {
	mc.FileName = name
}

func (mc *UploadedAttachmentInfo) SetMime(mime string) {
	mc.Mime = mime
}

func (mc *UploadedAttachmentInfo) SetSize(size int64) {
	mc.Size = size
}

func (mc *UploadedAttachmentInfo) SetRelativeUrl(url string) {
	mc.RelativeUrl = url
}
