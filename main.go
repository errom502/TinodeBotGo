package main

import (
	"csToGo25042023/CSbotToGo"
	"fmt"
)

type UserInfo struct {
	Username string `json:"username"`
	Passwd   string `json:"passwd"`
}

var bot CSbotToGo.ChatBot

type EventArgs struct{}
type ConsoleCancelEventArgs struct{}

func Bot_DisconnectedEvent(chan bool) {
	fmt.Println("[!!!Disconnected!!!]")
}

func Bot_LoginFailedEvent(sender interface{}, e CSbotToGo.EventArgs) {
	fmt.Println("[!!!Login Failed!!!]")
}

func Bot_LoginSuccessEvent(sender interface{}, e CSbotToGo.EventArgs) {
	fmt.Println("[!!!Login Success!!!]")
}

func Bot_CtrlMessageEvent(sender interface{}, e CSbotToGo.CtrlMessageEventArgs) {
	//fmt.Printf("[Ctrl Message] %s %s  %s  %s  %s  %s  %t\n", e.Code, e.Id, e.Topic, e.Text, e.Params, e.Type, e.HasError)
}

func Bot_ServerPresEvent(sender interface{}, e *CSbotToGo.ServerPresEventArgs) {
	//fmt.Printf("[Pres Message] %s\n", e.Pres.ToString())
}

func Bot_ServerMetaEvent(sender interface{}, e *CSbotToGo.ServerMetaEventArgs) {
	//fmt.Printf("[Meta Message] %s\n", e.Meta.ToString())
}

func Bot_ServerDataEvent(sender interface{}, e *CSbotToGo.ServerDataEventArgs) {
	//fmt.Printf("[Data Message] %s\n", e.Data.ToString())
}

func Console_CancelKeyPress(sender interface{}, e ConsoleCancelEventArgs) {
	bot.Stop()
}

type CmdOptions struct {
	CookieFile string `short:"C" long:"login-cookie" required:"false" default:".tn-cookie" help:"read credentials from the provided cookie file"`
	Token      string `short:"T" long:"login-token" required:"false" help:"login using token authentication"`
	Basic      string `short:"B" long:"login-basic" required:"false" help:"login using basic authentication username:password"`
	Listen     string `short:"L" long:"listen" required:"false" default:"0.0.0.0:40051" help:"address to listen on for incoming Plugin API calls"`
	Host       string `short:"S" long:"server" required:"false" default:"localhost:16060" help:"address of Tinode server gRPC endpoint"`
}

// переписать нормально main
func main() {

	//// Обработчик прерывания
	//signal.Notify(sigint, os.Interrupt)
	//bot.Chatbot("localhost:6061", "0.0.0.0:40052", ".tn-cookie", "basic", "YWxpY2U6YWxpY2UxMjM=", nil)
	//alice := "YWxpY2U6YWxpY2UxMjM="
	var host, listen, schemaArg, secretArg, cookieFile string
	//flag.StringVar(&host, "host", bot.ServerHost, "gRPC server")
	//flag.StringVar(&listen, "listen", bot.Listen, "Plugin API calls Listen server")
	//flag.StringVar(&schemaArg, "schema", bot.Schema, "Login schema (basic or token)")
	//flag.StringVar(&secretArg, "secret", bot.Secret, "Login secret")
	//flag.StringVar(&cookieFile, "cookie", bot.CookieFile, "Cookie file")
	//flag.Parse()

	var o CmdOptions
	o.Host = "localhost:16060"
	o.Listen = "0.0.0.0:40052"
	o.Basic = "alice:alice123"
	if o.Host != "" {
		host = o.Host
		fmt.Printf("gRPC server: %+v\n", host)
	}
	if o.Listen != "" {
		listen = o.Listen
		fmt.Printf("Plugin API calls Listen server: %+v\n", listen)
	}
	if o.Token != "" {
		schemaArg = "token"
		secretArg = string([]byte(o.Token))
		fmt.Printf("Login in with token %+v\n", o.Token)
		bot.Chatbot(host, listen, "", schemaArg, secretArg, nil)
	} else if o.Basic != "" {
		schemaArg = "basic"
		secretArg = string([]byte(o.Basic))
		fmt.Printf("Login in with login:password %+v\n", o.Basic)
		bot.Chatbot(host, listen, "", schemaArg, secretArg, nil)
	} else {
		cookieFile = o.CookieFile
		fmt.Printf("Login in with cookie file %+v\n", o.CookieFile)
		bot.Chatbot(host, listen, cookieFile, "", "", nil)

		outSchem, outSercet, ok := bot.ReadAuthCookie()
		if !ok {
			fmt.Printf("Login in with cookie file failed, please check your credentials and try again... Press any key to exit.")
			return
		} else {
			bot.Schema = outSchem
			bot.Secret = string(outSercet)
		}
	}
	bot.ServerDataEvent = Bot_ServerDataEvent
	bot.ServerMetaEvent = Bot_ServerMetaEvent
	bot.ServerPresEvent = Bot_ServerPresEvent
	bot.CtrlMessageEvent = Bot_CtrlMessageEvent
	bot.LoginSuccessEvent = Bot_LoginSuccessEvent
	bot.LoginFailedEvent = Bot_LoginFailedEvent
	//bot.DisconnectedEvent = Bot_DisconnectedEvent
	bot.SetHttpApi("http://localhost:6660", "AQAAAAABAABtfBKva9nJN3ykjBi0feyL")
	//bot.BotResponse.
	var b CSbotToGo.BotResponse
	b.BotResponse(bot)
	bot.BotResponse = b.IBotResponse
	bot.Start()

	//// Создание экземпляра ChatBot
	//switch {
	//case schemaArg == "token":
	//	var token []byte
	//	_, _, err := ascii85.Decode(token, []byte(secretArg), false)
	//	if err != nil {
	//		fmt.Println("Failed to decode token.")
	//		return
	//	}
	//	fmt.Printf("Login in with token %s\n", secretArg)
	//	bot.Chatbot(host, listen, cookieFile, schemaArg, secretArg, bot.BotResponse)
	//case schemaArg == "basic":
	//	fmt.Printf("Login in with login:password %s\n", secretArg)
	//	bot.Chatbot(host, listen, cookieFile, schemaArg, secretArg, bot.BotResponse)
	//default:
	//	if cookieFile == "" {
	//		fmt.Println("No authentication credentials provided.")
	//		return
	//	}
	//	fmt.Printf("Login in with cookie file %s\n", cookieFile)
	//	bot.Chatbot(host, listen, cookieFile, "", "", nil)
	//	schema, secret, err := bot.ReadAuthCookie(cookieFile)
	//	if !err {
	//		fmt.Printf("Failed to read cookie file: %v\n", err)
	//		return
	//	}
	//	bot.Schema = schema
	//	bot.Secret = string(secret)
	//}

	// Установка обработчиков событий
	//bot.ServerDataEvent = Bot_ServerDataEvent
	//bot.ServerMetaEvent = Bot_ServerMetaEvent
	//bot.ServerPresEvent = Bot_ServerPresEvent
	//bot.CtrlMessageEvent = Bot_CtrlMessageEvent
	//bot.LoginSuccessEvent = Bot_LoginSuccessEvent
	//bot.LoginFailedEvent = Bot_LoginFailedEvent
	//bot.DisconnectedEvent = Bot_DisconnectedEvent
	// Запуск ChatBot
	//fmt.Println("Starting ChatBot...")
	//fmt.Println("starting the golang chatbot.....")
	//// Set up a connection to the gRPC server.
	//conn, err := grpc.Dial(address, grpc.WithInsecure())
	//if err != nil {
	//	log.Fatalf("did not connect: %v", err)
	//}
	//defer conn.Close()
	//fmt.Println("server connected with:" + address)
	////
	//client := pb.NewNodeClient(conn)
	//nmc, err := client.MessageLoop(context.Background())
	//if err != nil {
	//	grpclog.Fatalln(err)
	//}
	// Установка HTTP API для обработки загрузки больших файлов

	// Ожидание прерывания
	//<-sigint

	// Остановка ChatBot
	//fmt.Println("[Bye Bye] ChatBot Stopped")
	//bot.Stop()
}
