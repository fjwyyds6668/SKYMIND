package models

import "fmt"

// PromptGenerator 提示词生成器定义
type PromptGenerator struct{}

// SystemPromptGenerator 生成系统提示词
func (PromptGenerator) SystemPromptGenerator(name, description, userInput string) string {
	var template string

	if userInput == "" {
		// 如果用户输入为空，则基于名称和描述生成新的系统提示词
		template = `你是一个专业的AI助手提示词优化专家。根据用户提供的助手信息，生成严谨、准确的系统提示词。

要求：
1. 提示词要以第二人称定义助手的角色和专业领域（如"你现在是一名..."）
2. 明确助手的专业技能、回答风格和行为准则
3. 添加必要的限制和安全考虑
4. 使用清晰、结构化的语言
5. 长度控制在50-200字之间
6. 重点是规定助手应该做什么，而不是向用户介绍自己

用户输入：
- 助手名称：` + name + `
- 助手描述：` + description + `
- 用户需求：请基于助手名称和描述生成一个专业、完整的系统提示词。

请直接返回优化后的系统提示词，不要包含其他解释。`
	} else {
		// 如果有用户输入，则基于现有信息优化系统提示词
		template = `你是一个专业的AI助手提示词优化专家。根据用户提供的助手信息，生成严谨、准确的系统提示词。

要求：
1. 提示词要以第二人称定义助手的角色和专业领域（如"你现在是一名..."）
2. 明确助手的专业技能、回答风格和行为准则
3. 添加必要的限制和安全考虑
4. 使用清晰、结构化的语言
5. 长度控制在50-200字之间
6. 重点是规定助手应该做什么，而不是向用户介绍自己

用户输入：
- 助手名称：` + name + `
- 助手描述：` + description + `
- 用户需求：` + userInput + `

请直接返回优化后的系统提示词，不要包含其他解释。`
	}

	return template
}

// UserPromptOptimizer 优化用户提示词
func (PromptGenerator) UserPromptOptimizer(originalPrompt string) string {
	template := `你是一个专业的提示词优化专家。帮助用户优化AI对话提示词，使其更加清晰、准确和有效。

优化原则：
1. 保持用户原意，提升表达清晰度
2. 添加必要的上下文信息
3. 使用更精确的词汇
4. 确保问题结构合理
5. 控制长度在合理范围内

原始提示词：` + originalPrompt + `

请直接返回优化后的用户提示词，不要包含其他解释，如果用户原提示词已经很好，可以返回原内容。`
	return template
}

// ConversationTitleGenerator 生成对话标题
func (PromptGenerator) ConversationTitleGenerator(userMessage, aiResponse string) string {
	template := `你是一个专业的对话内容总结专家。根据AI助手的回复内容，生成简洁、准确的对话标题。

要求：
1. 标题要概括对话的核心主题
2. 使用简洁明了的语言
3. 长度控制在10-20个字之间
4. 避免使用特殊字符和表情符号
5. 确保标题具有描述性和吸引力

用户输入内容：` + userMessage + `
AI回复内容：` + aiResponse + `

请直接返回生成的标题，不要包含其他内容，尤其是前导词`
	return template
}

// TopicTitleGenerator 生成话题标题
func (PromptGenerator) TopicTitleGenerator(conversationTitles []string) string {
	var titlesText string
	for i, title := range conversationTitles {
		if i > 0 {
			titlesText += "\n"
		}
		titlesText += fmt.Sprintf("%d. %s", i+1, title)
	}

	template := `你是一个专业的话题命名专家。根据一个话题下的多个对话标题，生成一个概括性和吸引力的话题标题。

要求：
1. 话题标题要比所有对话标题更加概括和抽象
2. 体现该话题下对话的共同主题和方向
3. 长度控制在8-15个字之间
4. 使用简洁、易懂的语言
5. 便于用户快速识别话题内容

对话标题列表：
` + titlesText + `

请直接返回生成的话题标题，不要包含其他内容，尤其是前导词`
	return template
}
