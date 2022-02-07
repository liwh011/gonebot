package gonebot

type State map[string]interface{}

func writePrefixToState(state State, prefix, raw string) {
	state["prefix"] = map[string]interface{}{
		"matched": prefix,
		"text":    raw[len(prefix):],
		"raw":     raw,
	}
}

func writeSuffixToState(state State, suffix, raw string) {
	state["suffix"] = map[string]interface{}{
		"matched": suffix,
		"text":    raw[:len(raw)-len(suffix)],
		"raw":     raw,
	}
}

func writeCommandToState(state State, rawCmd, cmd, raw string) {
	state["command"] = map[string]interface{}{
		"matched": cmd,
		"raw_cmd": rawCmd,
		"text":    raw[len(rawCmd):],
		"raw":     raw,
	}
}

func writeKeywordToState(state State, keyword string) {
	state["keyword"] = map[string]interface{}{
		"matched": keyword,
	}
}

func writeRegexToState(state State, matchGroup []string) {
	state["regex"] = map[string]interface{}{
		"matched": matchGroup,
	}
}
