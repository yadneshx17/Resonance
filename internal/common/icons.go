package common

type Style struct {
	Icon  string
	Color string
}

var (
	Directory = "\uf07b"  // Printable Rune : ""
	Music     = ""       // Printable Rune : ""
	Cursor    = "\uf054 " // Printable Rune : ""

	// Media Controls
	Play        = "\uf04b"     // Printable Rune : ""
	Pause       = "\uf04c"     // Printable Rune : ""
	Stop        = "\uf04d"     // Printable Rune : ""
	FastForward = "\uf04e"     // Printable Rune : ""
	Rewind      = "\uf04a"     // Printable Rune : ""
	NextTrack   = "\uf051"     // Printable Rune : ""
	PrevTrack   = "\uf048"     // Printable Rune : ""
	Shuffle     = "\uf074"     // Printable Rune : ""
	Repeat      = "\uf01e"     // Printable Rune : ""
	RepeatOne   = "\U000f0458" // Printable Rune : "󰑘"
	Eject       = "\uf052"     // Printable Rune : ""

	// Volume & Output
	VolumeMute   = "\uf6a9"     // Printable Rune : ""
	VolumeLow    = "\uf027"     // Printable Rune : ""
	VolumeMedium = "\U000f057e" // Printable Rune : "󰕾"
	VolumeHigh   = "\uf028"     // Printable Rune : ""
	Headphones   = "\uf025"     // Printable Rune : ""
	Speaker      = "\U000f04ca" // Printable Rune : "󰓊"
	Bluetooth    = "\uf293"     // Printable Rune : ""
	Cast         = "\U000f0122" // Printable Rune : "󰄢"

	// Audio Library
	MusicNote  = "\uf001"     // Printable Rune : ""
	MusicNotes = "\uf886"     // Printable Rune : ""
	Album      = "\U000f1362" // Printable Rune : "󱍢"
	Disc       = "\uf10c"     // Printable Rune : ""
	Playlist   = "\U000f0c0c" // Printable Rune : "󰰌"
	Guitar     = "\U000f0c97" // Printable Rune : "󰲗"
	Piano      = "\U000f074d" // Printable Rune : "󰝍"
	Drum       = "\U000f0f29" // Printable Rune : "󰼩"

	// Recording & Gear
	Microphone = "\uf130"     // Printable Rune : ""
	MicMuted   = "\U000f036d" // Printable Rune : "󰍭"
	Waveform   = "\U000f0724" // Printable Rune : "󰜤"
	Equalizer  = "\uf3f6"     // Printable Rune : ""
	Radio      = "\uf57f"     // Printable Rune : ""
	Metronome  = "\U000f10c0" // Printable Rune : "󱃀"
	Podcast    = "\uf2ce"     // Printable Rune : ""

	// Library Interaction
	LikeTrack   = "\uf004"     // Printable Rune : ""
	UnlikeTrack = "\uf08a"     // Printable Rune : ""
	QueueAdd    = "\U000f0612" // Printable Rune : "󰘒"
	QueueRemove = "\U000f0d2c" // Printable Rune : "󰴬"
	Lyrics      = "\U000f0367" // Printable Rune : "󰍧"
	SearchMusic = "\U000f1883" // Printable Rune : "󱢃"

	Search = "\ue68f" // Printable Rune : ""
)

var Icons = map[string]Style{
	"audio": {Icon: "\uf001", Color: "#ee524f"}, // Printable Rune : ""
}
