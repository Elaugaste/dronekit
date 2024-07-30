package modem

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type LinkState struct {
	Mode     string
	Strength string
	LQ       string
}

func (s LinkState) String() string {
	return fmt.Sprintf("%s PWR:%s LQ:%s", s.Mode, s.Strength, s.LQ)
}

func getMode() string {
	client.Escape()
	r, err := client.Command("+COPS?")
	if err != nil {
		fmt.Printf("COPS Error: %T\n", err)
		return "ERR"
	}

	comp, _ := regexp.Compile("\\+COPS:.*,[0-9]")
	q := comp.FindString(r[0])
	if q == "" {
		return "ERR"
	}
	q = q[len(q)-1:]
	mode, err := strconv.Atoi(q)
	if err != nil {
		fmt.Printf("Atoi Error: %v\n", r)
		return "ERR"
	}

	return mode2text(mode)
}

func getLQ() []string {

	//+CSQ returns 2 values separated by a comma.
	//The first value represents the signal strength and provides a value between 0 and 31;
	//higher numbers indicate better signal strength.
	//The second value represents the signal quality indicated by a value between 0 and 7.
	//If AT+CSQ returns 99,99, the signal is undetectable or unknown.
	client.Escape()
	r, err := client.Command("+CSQ")
	if err != nil {
		fmt.Printf("CSQ Error: %T\n", err)
		return []string{"ERR", "ERR"}
	}

	if r[0][6:] == "99,99" {
		fmt.Println("incorrect link levels")
		return []string{"ERR", "ERR"}
	}

	q := strings.Split(r[0][6:], ",")
	if len(q) != 2 {
		fmt.Printf("Split Error: %s\n", r[0])
		return []string{"ERR", "ERR"}
	}

	strength, _ := strconv.Atoi(q[0])
	lq, _ := strconv.Atoi(q[1])

	return []string{level2string(strength, 31), level2string(lq, 7)}
}

func level2string(lvl, maxlvl int) string {
	if lvl <= maxlvl/3 {
		return "LOW"
	} else if lvl <= (maxlvl/3)*2 {
		return "MED"
	}
	return "HIGH"
}
