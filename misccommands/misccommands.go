package misccommands

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/ThyLeader/rikka"
	"github.com/go-redis/redis"
)

var pepe = `:frog::frog::frog::frog::frog::frog::frog:
_:frog::frog::frog::frog::frog::frog::frog::frog::frog:
__:frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog:
:frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog:
:frog::frog::white_circle::black_circle::black_circle::white_circle::frog::frog::frog::white_circle::black_circle::black_circle::white_circle:
:frog::white_circle::black_circle::black_circle::white_circle::black_circle::white_circle::frog::white_circle::black_circle::black_circle::white_circle::black_circle::white_circle:
:frog::white_circle::black_circle::white_circle::black_circle::black_circle::white_circle::frog::white_circle::black_circle::white_circle::black_circle::black_circle::white_circle:
:frog::frog::white_circle::black_circle::white_circle::white_circle::frog::frog::frog::white_circle::black_circle::white_circle::white_circle:
:frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog:
:red_circle::red_circle::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog:
:frog::red_circle::red_circle::frog::frog::frog::frog::frog::frog::frog::frog::frog:
:frog::frog::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle:
:frog::frog::frog::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle:
:frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog:
:frog::frog::frog::frog::frog::frog::frog::frog::frog::frog:
:frog::frog::frog::frog::frog::frog::frog::frog::frog:`

var client = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

func MessagePeepo(bot *rikka.Bot, service rikka.Service, message rikka.Message, command string, parts []string) {
	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "pepe", message) {
		return
	}
	service.SendMessage(message.Channel(), pepe)
}

var HelpPeepo = rikka.NewCommandHelp("", "Sends a pepe.")

var userIDRegex = regexp.MustCompile("<@!?([0-9]*)>")
var chanIDRegex = regexp.MustCompile("<#!?([0-9]*)>")

func MessageIDTS(bot *rikka.Bot, service rikka.Service, message rikka.Message, command string, parts []string) {
	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "ts", message) {
		return
	}

	query := strings.Join(strings.Split(message.RawMessage(), " ")[1:], " ")
	var id string
	if len(parts) == 0 {
		id = message.UserID()
	} else {
		if q := userIDRegex.FindStringSubmatch(query); q != nil {
			id = q[1]
		}
		if q := chanIDRegex.FindStringSubmatch(query); q != nil {
			id = q[1]
		}
		if id == "" {
			id = parts[0]
		}
	}
	t, err := service.TimestampForID(id)
	if err != nil {
		service.SendMessage(message.Channel(), "Incorrect snowflake")
		return
	}
	service.SendMessage(message.Channel(), fmt.Sprintf("`%s`", t.UTC().Format(time.UnixDate)))
}

// HelpIDTS is the help function for timestamp parsing
var HelpIDTS = rikka.NewCommandHelp("[@username]", "Parses a snowflake (id) and returns a timestamp.")

// MessageSupport is the message handler for support
func MessageSupport(bot *rikka.Bot, service rikka.Service, message rikka.Message, command string, parts []string) {
	service.SendMessage(message.Channel(), "You can join the support server here: https://rikka.xyz")
}

// HelpSupport is the help function for support
var HelpSupport = rikka.NewCommandHelp("", "Gives an invite link to join the support server.")

// MessagePing is the command handler for the ping command
func MessagePing(bot *rikka.Bot, service rikka.Service, message rikka.Message, command string, parts []string) {
	now := time.Now()
	p, _ := service.SendMessage(message.Channel(), "Pong!")
	after := time.Now()

	service.EditMessage(message.Channel(), p.ID, fmt.Sprintf("Pong! - `%s`", after.Sub(now).String()))
}

// HelpPing is the help text for the ping command
var HelpPing = rikka.NewCommandHelp("", "Shows bot latency.")

// MessageExclude excludes people from using the bot
func MessageExclude(bot *rikka.Bot, service rikka.Service, message rikka.Message, command string, parts []string) {
	if !service.IsBotOwner(message) {
		service.SendMessage(message.Channel(), "Sorry, you must be the owner to use this command")
		return
	}
	if len(parts) != 1 {
		service.SendMessage(message.Channel(), "userid not provided")
		return
	}
	err := client.SAdd("exclude", parts[0]).Err()
	if err != nil {
		service.SendMessage(message.Channel(), err.Error())
		return
	}
	service.SendMessage(message.Channel(), fmt.Sprintf("Successfully excluded user `%s`", parts[0]))
}

// MessageUnexclude excludes people from using the bot
func MessageUnexclude(bot *rikka.Bot, service rikka.Service, message rikka.Message, command string, parts []string) {
	if !service.IsBotOwner(message) {
		service.SendMessage(message.Channel(), "Sorry, you must be the owner to use this command")
		return
	}
	if len(parts) != 1 {
		service.SendMessage(message.Channel(), "userid not provided")
		return
	}
	err := client.SRem("exclude", parts[0]).Err()
	if err != nil {
		service.SendMessage(message.Channel(), err.Error())
		return
	}
	service.SendMessage(message.Channel(), fmt.Sprintf("Successfully unexcluded user `%s`", parts[0]))
}

// MessageLenny is the handler for the lenny command
func MessageLenny(bot *rikka.Bot, service rikka.Service, message rikka.Message, command string, parts []string) {
	r, _ := rand.Int(rand.Reader, big.NewInt(int64(len(lenny))))
	service.SendMessage(message.Channel(), lenny[int(r.Int64())])
}

// HelpLenny is the help text for the lenny command
var HelpLenny = rikka.NewCommandHelp("", "Sends a random lenny")

var lenny = []string{
	"( ͡° ͜ʖ ͡°)",
	"( ͡°( ͡° ͜ʖ( ͡° ͜ʖ ͡°)ʖ ͡°) ͡°)",
	"┬┴┬┴┤ ͜ʖ ͡°) ├┬┴┬┴",
	"[̲̅$̲̅(̲̅ ͡° ͜ʖ ͡°̲̅)̲̅$̲̅]",
	"( ͡ᵔ ͜ʖ ͡ᵔ )",
	"༼ つ  ͡° ͜ʖ ͡° ༽つ",
	"Please appear before us, oh mighty soldier of the dark!\n╰( ͡° ͜ʖ ͡° )つ──☆ﾟ.･｡ﾟ      ☆･PUFF･ *☆      ^(#｀∀´)_Ψ",
	"( .-. )",
	"( .o.)",
	"( `·´ )",
	"( ° ͜ ʖ °)",
	"( ͡° ͜ʖ ͡°)",
	"( ⚆ _ ⚆ )",
	"( ︶︿︶)",
	"( ﾟヮﾟ)",
	"(\\/)(°,,,°)(\\/)",
	"(¬_¬)",
	"(¬º-°)¬",
	"(¬‿¬)",
	"(°ロ°)☝",
	"(´・ω・)っ",
	"(ó ì_í)",
	"(ʘᗩʘ')",
	"(ʘ‿ʘ)",
	"(̿▀̿ ̿Ĺ̯̿̿▀̿ ̿)̄",
	"(͡° ͜ʖ ͡°)",
	"ᕦ( ͡° ͜ʖ ͡°)ᕤ",
	"(ಠ_ಠ)",
	"(ಠ‿ಠ)",
	"(ಠ⌣ಠ)",
	"(ಥ_ಥ)",
	"(ಥ﹏ಥ)",
	"(ง ͠° ͟ل͜ ͡°)ง",
	"(ง ͡ʘ ͜ʖ ͡ʘ)ง",
	"(ง •̀_•́)ง",
	"(ง'̀-'́)ง",
	"(ง°ل͜°)ง",
	"(ღ˘⌣˘ღ)",
	"(ᵔᴥᵔ)",
	"(•ω•)",
	"(•◡•)/",
	"(⊙ω⊙)",
	"(⌐■_■)",
	"(─‿‿─)",
	"(╯°□°）╯",
	"(◕‿◕)",
	"(☞ﾟ∀ﾟ)☞",
	"(❍ᴥ❍ʋ)",
	"(っ◕‿◕)っ",
	"(づ｡◕‿‿◕｡)づ",
	"(ノಠ益ಠ)ノ",
	"(ノ・∀・)ノ",
	"(；一_一)",
	"(｀◔ ω ◔´)",
	"(｡◕‿‿◕｡)",
	"(ﾉ◕ヮ◕)ﾉ",
	"*<{:¬{D}}}",
	"=^.^=",
	"t(-.-t)",
	"| (• ◡•)|",
	"~(˘▾˘~)",
	"¬_¬",
	"¯(°_o)/¯",
	"¯\\_(ツ)_/¯",
	"°Д°",
	"ɳ༼ຈل͜ຈ༽ɲ",
	"ʅʕ•ᴥ•ʔʃ",
	"ʕ´•ᴥ•`ʔ",
	"ʕ•ᴥ•ʔ",
	"ʕ◉.◉ʔ",
	"ʕㅇ호ㅇʔ",
	"ʕ；•`ᴥ•´ʔ",
	"ʘ‿ʘ",
	"͡° ͜ʖ ͡°",
	"ζ༼Ɵ͆ل͜Ɵ͆༽ᶘ",
	"Ѱζ༼ᴼل͜ᴼ༽ᶘѰ",
	"ب_ب",
	"٩◔̯◔۶",
	"ಠ_ಠ",
	"ಠoಠ",
	"ಠ~ಠ",
	"ಠ‿ಠ",
	"ಠ⌣ಠ",
	"ಠ╭╮ಠ",
	"ರ_ರ",
	"ง ͠° ل͜ °)ง",
	"༼ ºººººل͟ººººº ༽",
	"༼ ºل͟º ༽",
	"༼ ºل͟º༼",
	"༼ ºل͟º༽",
	"༼ ͡■ل͜ ͡■༽",
	"༼ つ ◕_◕ ༽つ",
	"༼ʘ̚ل͜ʘ̚༽",
	"ლ(´ڡ`ლ)",
	"ლ(́◉◞౪◟◉‵ლ)",
	"ლ(ಠ益ಠლ)",
	"ᄽὁȍ ̪őὀᄿ",
	"ᔑ•ﺪ͟͠•ᔐ",
	"ᕕ( ᐛ )ᕗ",
	"ᕙ(⇀‸↼‶)ᕗ",
	"ᕙ༼ຈل͜ຈ༽ᕗ",
	"ᶘ ᵒᴥᵒᶅ",
	"(ﾉಥ益ಥ）ﾉ",
	"≧☉_☉≦",
	"⊙▃⊙",
	"⊙﹏⊙",
	"┌( ಠ_ಠ)┘",
	"╚(ಠ_ಠ)=┐",
	"◉_◉",
	"◔ ⌣ ◔",
	"◔̯◔",
	"◕‿↼",
	"◕‿◕",
	"☉_☉",
	"☜(⌒▽⌒)☞",
	"☼.☼",
	"♥‿♥",
	"⚆ _ ⚆",
	"✌(-‿-)✌",
	"〆(・∀・＠)",
	"ノ( º _ ºノ)",
	"ノ( ゜-゜ノ)",
	"ヽ( ͝° ͜ʖ͡°)ﾉ",
	"ヽ(`Д´)ﾉ",
	"ヽ༼° ͟ل͜ ͡°༽ﾉ",
	"ヽ༼ʘ̚ل͜ʘ̚༽ﾉ",
	"ヽ༼ຈل͜ຈ༽ง",
	"ヽ༼ຈل͜ຈ༽ﾉ",
	"ヽ༼Ὸل͜ຈ༽ﾉ",
	"ヾ(⌐■_■)ノ",
	"꒰･◡･๑꒱",
	"｡◕‿◕｡",
	"ʕノ◔ϖ◔ʔノ",
	"(ノಠ益ಠ)ノ彡┻━┻",
	"(╯°□°）╯︵ ┻━┻",
	"꒰•̥̥̥̥̥̥̥ ﹏ •̥̥̥̥̥̥̥̥๑꒱",
	"ಠ_ರೃ",
	"(ू˃̣̣̣̣̣̣︿˂̣̣̣̣̣̣ ू)",
	"(ꈨຶꎁꈨຶ)۶”",
	"(ꐦ°᷄д°᷅)",
	"٩꒰･ัε･ั ꒱۶",
	"ヘ（。□°）ヘ",
	"˓˓(ृ　 ु ॑꒳’)ु(ृ’꒳ ॑ ृ　)ु˒˒˒",
	"꒰✘Д✘◍꒱",
	"૮( ᵒ̌ૢཪᵒ̌ૢ )ა",
	"“ψ(｀∇´)ψ",
	"ಠﭛಠ",
	"(๑>ᴗ<๑)",
	"(۶ꈨຶꎁꈨຶ )۶ʸᵉᵃʰᵎ",
	"٩(•̤̀ᵕ•̤́๑)ᵒᵏᵎᵎᵎᵎ",
	"(oT-T)尸",
	"ಥ‿ಥ",
	"┬┴┬┴┤  (ಠ├┬┴┬┴",
	"( ˘ ³˘)♥",
	"Σ (੭ु ຶਊ ຶ)੭ु⁾⁾",
	"(⑅ ॣ•͈ᴗ•͈ ॣ)",
	"ヾ(´￢｀)ﾉ",
	"(•̀o•́)ง",
	"=͟͟͞͞ =͟͟͞͞ ﾍ( ´Д`)ﾉ",
	"(((╹д╹;)))",
	"•̀.̫•́✧",
	"(ᵒ̤̑ ₀̑ ᵒ̤̑)",
	"\\_(ʘ_ʘ)_/",
	"乙(ツ)乙",
	"乙(のっの)乙",
	"ヾ(¯∇￣๑)",
	"\\_(ʘ_ʘ)_/",
	"༼;´༎ຶ ۝ ༎ຶ༽",
	"(▀̿Ĺ̯▀̿ ̿)",
	"(ﾉ◕ヮ◕)ﾉ*:･ﾟ✧",
	"┬┴┬┴┤ ͜ʖ ͡°) ├┬┴┬┴",
	"┬┴┬┴┤(･_├┬┴┬┴",
	"(͡ ͡° ͜ つ ͡͡°)",
	"( ͡°╭͜ʖ╮͡° )",
	"(• ε •)",
	"[̲̅$̲̅(̲̅ ͡° ͜ʖ ͡°̲̅)̲̅$̲̅]",
	"| (• ◡•)| (❍ᴥ❍ʋ)",
	"(◕‿◕✿)",
	"(╯°□°)╯︵ ʞooqǝɔɐɟ",
	"(☞ﾟヮﾟ)☞ ☜(ﾟヮﾟ☜)",
	"(づ￣ ³￣)づ",
	"(;´༎ຶД༎ຶ`)",
	"♪~ ᕕ(ᐛ)ᕗ",
	"༼ つ  ͡° ͜ʖ ͡° ༽つ",
	"༼ つ ಥ_ಥ ༽つ",
	"ಥ_ಥ",
	"( ͡ᵔ ͜ʖ ͡ᵔ )",
	"ヾ(⌐■_■)ノ♪",
	"~(˘▾˘~)",
	"\\ (•◡•) /",
	"(~˘▾˘)~",
	"(._.) ( l: ) ( .-. ) ( :l ) (._.)",
	"༼ ºل͟º ༼ ºل͟º ༼ ºل͟º ༽ ºل͟º ༽ ºل͟º ༽",
	"┻━┻ ︵ヽ(`Д´)ﾉ︵ ┻━┻",
	"ᕦ(ò_óˇ)ᕤ",
	"(•_•) ( •_•)>⌐■-■ (⌐■_■)",
	"(☞ຈل͜ຈ)☞",
	"˙ ͜ʟ˙",
	"☜(˚▽˚)☞",
	"(｡◕‿◕｡)",
	"（╯°□°）╯︵( .o.)",
	"(っ˘ڡ˘ς)",
	"┬──┬ ノ( ゜-゜ノ)",
	"ಠ⌣ಠ",
	"( ಠ ͜ʖರೃ)",
	"ƪ(˘⌣˘)ʃ",
	"¯\\(°_o)/¯",
	"ლ,ᔑ•ﺪ͟͠•ᔐ.ლ",
	"(´・ω・`)",
	"(´・ω・)っ由",
	"(° ͡ ͜ ͡ʖ ͡ °)",
	"Ƹ̵̡Ӝ̵̨̄Ʒ",
	"ಠ_ಥ",
	"ಠ‿↼",
	"(>ლ)",
	"(▰˘◡˘▰)",
	"(✿´‿`)",
	"◔ ⌣ ◔",
	"｡゜(｀Д´)゜｡",
	"┬─┬ノ( º _ ºノ)",
	"(ó ì_í)=óò=(ì_í ò)",
	"(/) (°,,°) (/)",
	"┬─┬ ︵ /(.□. ）",
	"^̮^",
	"(>人<)",
	"(~_^)",
	"(･.◤)",
	">_>",
	"(^̮^)",
	"=U",
	"(｡╹ω╹｡)",
	"ლ(╹◡╹ლ)",
	"(●´⌓`●)",
	"（[∂]ω[∂]）",
	"U^ｴ^U",
	"(〒ó〒)",
	"(T^T)",
	"(íoì)",
	"(#•v•#)",
	"(•^u^•)",
	"!(^3^)!",
	"\\(°°\\”)",
	"(°o°:)",
	"(° o°)!",
	"(oﾛo)!!",
	"(òロó)",
	"(ò皿ó)",
	"(￣･_______･￣)",
	"ヾ(๑╹◡╹)ﾉ'",
	"(ლ╹◡╹)ლ",
	"（◞‸◟）",
	"(✿◖◡◗)",
	"(　´･‿･｀)",
	"(*｀益´*)がう",
	"(ヾﾉ'д'o)ﾅｨﾅｨ",
	"❤(◕‿◕✿)",
	"(◡‿◡*)❤",
	"(o'ω'o)",
	"(｡･ˇ_ˇ･｡)ﾑｩ…",
	"♬♩♫♪☻(●´∀｀●）☺♪♫♩♬",
	"(つД⊂)ｴｰﾝ",
	"(つД・)ﾁﾗ",
	"(*´ω｀*)",
	"(✪‿✪)ノ",
	"╲(｡◕‿◕｡)╱",
	"ლ(^o^ლ)",
	"https://www.tenor.co/FzFX.gif",
}
