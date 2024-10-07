package badge

type Badge struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color,omitempty"`
	LabelColor    string `json:"labelColor,omitempty"`
	IsError       bool   `json:"isError,omitempty"`
	NamedLogo     string `json:"namedLogo,omitempty"`
	LogoSvg       string `json:"logoSvg,omitempty"`
	LogoColor     string `json:"logoColor,omitempty"`
	LogoWidth     string `json:"logoWidth,omitempty"`
	Style         string `json:"style,omitempty"`
}

type Type string

const (
	Success Type = "success"
	Warning Type = "warning"
	Error   Type = "error"
)

func NewDependencyBadge(message string, badgeType Type, theme string) *Badge {
	isError := badgeType == Error || badgeType == Warning

	if theme == "renovate" {
		return &Badge{
			SchemaVersion: 1,
			Label:         "\t",
			LabelColor:    "2a2f64",
			LogoSvg:       string(reproducibleBuildsLogo),
			Message:       message,
			Color:         getColor(badgeType, theme),
			IsError:       isError,
		}
	} else if theme == "dependabot" {
		return &Badge{
			SchemaVersion: 1,
			Label:         "Reproducible Builds",
			LabelColor:    "2a2f64",
			LogoSvg:       string(reproducibleBuildsLogo),
			Message:       message,
			Color:         getColor(badgeType, theme),
			IsError:       isError,
		}
	}

	return &Badge{
		SchemaVersion: 1,
		Label:         "Reproducible Builds",
		LabelColor:    "2a2f64",
		LogoSvg:       string(reproducibleBuildsLogo),
		Message:       message,
		Color:         getColor(badgeType, theme),
		IsError:       isError,
	}
}

func getColor(badgeType Type, theme string) string {
	switch badgeType {
	case Success:
		return "brightgreen"
	case Warning:
		return "orangered"
	case Error:
		return "crimson"
	default:
		return "lightgrey"
	}
}
