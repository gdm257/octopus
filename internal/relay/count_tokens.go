package relay

import (
	"encoding/json"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/looplj/axonhub/llm/httpclient"
	"github.com/looplj/axonhub/llm/transformer/anthropic"
)

// charsPerToken 为字符到 token 的换算密度, 官方近似 3.5~4 字符每 token, 取下界使估算略偏高
const charsPerToken = 3.5

// imageBlockTokens 为图片与文档内容块的固定估算成本, 其 base64 数据不计入字符统计
const imageBlockTokens = 1600

// CountTokens 本地估算 Anthropic Messages 请求的输入 token 数并按协议返回。
// 客户端如 Claude Code 依赖该端点统计上下文占用, 端点缺失时会退化为对每个上下文条目
// 各发送一次 max_tokens 为 1 的真实推理请求; 本地估算不选择渠道, 也不产生上游用量,
// 因此不校验模型对应的分组是否存在。
func CountTokens(c *gin.Context) {
	inbound := anthropic.NewInboundTransformer()
	raw, err := httpclient.ReadHTTPRequest(c.Request)
	if err != nil {
		rejectRequest(c, inbound, err)
		return
	}
	var payload any
	if err := json.Unmarshal(raw.Body, &payload); err != nil {
		rejectRequest(c, inbound, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"input_tokens": estimateTokens(payload)})
}

// estimateTokens 递归累计请求正文中全部字符串值并按字符密度换算为 token 数。
// JSON 的键, 数字与布尔值不参与统计, 需要计数的内容全部落在字符串值里;
// 图片与文档块没有字符串值可循, 按固定成本估算。
func estimateTokens(node any) int {
	switch value := node.(type) {
	case string:
		return int(math.Ceil(float64(len(value)) / charsPerToken))
	case []any:
		total := 0
		for _, item := range value {
			total += estimateTokens(item)
		}
		return total
	case map[string]any:
		if value["type"] == "image" || value["type"] == "document" {
			return imageBlockTokens
		}
		total := 0
		for _, item := range value {
			total += estimateTokens(item)
		}
		return total
	default:
		return 0
	}
}
