package modem

const (
	GSM         = 0
	GSM_Compact = 1
	UTRAN       = 2
	LTE         = 7
	CDMA        = 8
)

func mode2text(n int) string {
	switch n {
	case GSM:
		return "GSM"
	case GSM_Compact:
		return "GSM_Compact"
	case UTRAN:
		return "UTRAN"
	case LTE:
		return "LTE"
	case CDMA:
		return "CDMA"
	}
	return "ERR4"
}
