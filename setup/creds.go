package setup

// PresetApp holds a pre-registered Telegram application credential.
type PresetApp struct {
	AppID   int    `json:"app_id"`
	AppHash string `json:"api_hash"`
}

// Built-in presets so users don't have to register their own app.
var Presets = map[string]PresetApp{
	"desktop": {
		// Telegram Desktop official client credentials.
		// https://github.com/telegramdesktop/tdesktop/blob/dev/Telegram/SourceFiles/config.h
		AppID:   2040,
		AppHash: "b18441a1ff607e10a989891a5462e627",
	},
	"tdl": {
		// Application registered by iyear, author of tdl.
		// https://github.com/iyear/tdl
		AppID:   15055931,
		AppHash: "021d433426cbb920eeb95164498fe3d3",
	},
}

// DefaultPreset is the default credential source for new users.
const DefaultPreset = "desktop"
