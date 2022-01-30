package handler

type State map[string]interface{}

type StatePrefix State

func writePrefixToState(state State, prefix, raw string) {
	state["_prefix"] = map[string]interface{}{
		"matched": prefix,
		"text":    raw[len(prefix):],
		"raw":     raw,
	}
}

// 获取匹配到的前缀
func (s StatePrefix) GetMatchedPrefix() string {
	return s["_prefix"].(map[string]interface{})["matched"].(string)
}

// 获取剩余的文本
func (s StatePrefix) GetTextWithoutPrefix() string {
	return s["_prefix"].(map[string]interface{})["text"].(string)
}

// 获取包含前缀的原始的文本
func (s StatePrefix) GetRawText() string {
	return s["_prefix"].(map[string]interface{})["raw"].(string)
}

type StateSuffix State

func writeSuffixToState(state State, suffix, raw string) {
	state["_suffix"] = map[string]interface{}{
		"matched": suffix,
		"text":    raw[:len(raw)-len(suffix)],
		"raw":     raw,
	}
}

// 获取匹配到的后缀
func (s StateSuffix) GetMatchedSuffix() string {
	return s["_suffix"].(map[string]interface{})["matched"].(string)
}

// 获取剩余的文本
func (s StateSuffix) GetTextWithoutSuffix() string {
	return s["_suffix"].(map[string]interface{})["text"].(string)
}

// 获取包含后缀的原始的文本
func (s StateSuffix) GetRawText() string {
	return s["_suffix"].(map[string]interface{})["raw"].(string)
}

type StateCommand State

func writeCommandToState(state State, rawCmd, cmd, raw string) {
	state["_command"] = map[string]interface{}{
		"matched": cmd,
		"raw_cmd": rawCmd,
		"text":    raw[len(rawCmd):],
		"raw":     raw,
	}
}

// 获取匹配到的命令
func (s StateCommand) GetMatchedCommand() string {
	return s["_command"].(map[string]interface{})["matched"].(string)
}

// 获取剩余的文本
func (s StateCommand) GetTextWithoutCommand() string {
	return s["_command"].(map[string]interface{})["text"].(string)
}

// 获取包含命令的原始的文本
func (s StateCommand) GetRawText() string {
	return s["_command"].(map[string]interface{})["raw"].(string)
}

type StateKeyword State

func writeKeywordToState(state State, keyword string) {
	state["_keyword"] = map[string]interface{}{
		"matched": keyword,
	}
}

// 获取匹配到的关键词
func (s StateKeyword) GetMatchedKeyword() string {
	return s["_keyword"].(map[string]interface{})["matched"].(string)
}

type StateRegex State

func writeRegexToState(state State, matchGroup []string) {
	state["_regex"] = map[string]interface{}{
		"matched": matchGroup,
	}
}

func (s StateRegex) GetMatchedGroup() []string {
	return s["_regex"].(map[string]interface{})["matched"].([]string)
}
