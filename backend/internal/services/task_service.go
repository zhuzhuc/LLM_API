package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// TaskService 专门处理特定任务的服务
type TaskService struct {
	modelManager *ModelManager
}

// NewTaskService 创建新的任务服务
func NewTaskService(modelManager *ModelManager) *TaskService {
	return &TaskService{
		modelManager: modelManager,
	}
}

// FileFormatRequest 文件格式转换请求
type FileFormatRequest struct {
	SourceFormat string                 `json:"source_format"`
	TargetFormat string                 `json:"target_format"`
	Content      string                 `json:"content"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

// FileFormatResponse 文件格式转换响应
type FileFormatResponse struct {
	ConvertedContent string `json:"converted_content"`
	Success          bool   `json:"success"`
	Message          string `json:"message,omitempty"`
}

// HomeworkRequest 作业批改请求
type HomeworkRequest struct {
	Subject    string `json:"subject"`
	Question   string `json:"question"`
	Answer     string `json:"answer"`
	GradeLevel string `json:"grade_level,omitempty"`
	Language   string `json:"language,omitempty"`
}

// HomeworkResponse 作业批改响应
type HomeworkResponse struct {
	Score       int      `json:"score"`
	Feedback    string   `json:"feedback"`
	Suggestions []string `json:"suggestions"`
	Success     bool     `json:"success"`
	Message     string   `json:"message,omitempty"`
}

// SubtitleRequest 字幕处理请求
type SubtitleRequest struct {
	VideoPath    string `json:"video_path"`
	SourceLang   string `json:"source_lang"`
	TargetLang   string `json:"target_lang"`
	OutputFormat string `json:"output_format,omitempty"`
}

// SubtitleResponse 字幕处理响应
type SubtitleResponse struct {
	SubtitlePath string `json:"subtitle_path"`
	Content      string `json:"content"`
	Success      bool   `json:"success"`
	Message      string `json:"message,omitempty"`
}

// ConvertFileFormat 文件格式转换
func (ts *TaskService) ConvertFileFormat(ctx context.Context, req FileFormatRequest) (*FileFormatResponse, error) {
	// 构建专门的提示词
	prompt := fmt.Sprintf(`你是一个专业的文件格式转换专家。请严格按照要求将以下内容进行格式转换。

**任务**: 将 %s 格式转换为 %s 格式

**源内容**:
%s

**转换要求**:
1. 严格保持数据结构和内容完整性
2. 遵循 %s 格式的标准语法规范
3. 保持数据类型的准确性（字符串、数字、布尔值、数组等）
4. 确保转换后的格式可以被标准解析器正确解析
5. 不要添加任何解释或额外内容

**输出**: 请直接输出转换后的 %s 格式内容，不要包含任何其他文字：`, req.SourceFormat, req.TargetFormat, req.Content, req.TargetFormat, req.TargetFormat)

	// 使用专用的格式转换模型
	response, err := ts.modelManager.ChatWithModel(ctx, "deepseek-coder-1.3b-format", prompt, 2048)
	if err != nil {
		return &FileFormatResponse{
			Success: false,
			Message: fmt.Sprintf("格式转换失败: %v", err),
		}, err
	}

	return &FileFormatResponse{
		ConvertedContent: response,
		Success:          true,
	}, nil
}

// GradeHomework 作业批改
func (ts *TaskService) GradeHomework(ctx context.Context, req HomeworkRequest) (*HomeworkResponse, error) {
	language := req.Language
	if language == "" {
		language = "中文"
	}

	prompt := fmt.Sprintf(`你是一名经验丰富的%s老师，请认真批改以下%s年级的作业。

**作业信息**:
- 科目: %s
- 年级: %s
- 题目: %s
- 学生答案: %s

**批改要求**:
1. 仔细分析题目要求和学生答案
2. 根据答案的正确性、完整性、逻辑性进行评分
3. 提供建设性的反馈和改进建议
4. 使用%s进行回复

**请按以下JSON格式输出批改结果**:
{
  "score": 分数(0-100的整数),
  "feedback": "详细的批改反馈，包括答案分析",
  "suggestions": "具体的改进建议和学习指导",
  "correct_answer": "如果学生答案有误，请提供正确答案或解题思路"
}

请确保输出是有效的JSON格式：`, req.Subject, req.GradeLevel, req.Subject, req.GradeLevel, req.Question, req.Answer, language)

	// 使用专用的教学模型
	response, err := ts.modelManager.ChatWithModel(ctx, "qwen2-7b-teacher", prompt, 1024)
	if err != nil {
		return &HomeworkResponse{
			Success: false,
			Message: fmt.Sprintf("作业批改失败: %v", err),
		}, err
	}

	// 尝试解析JSON格式的响应
	var jsonResponse struct {
		Score         int    `json:"score"`
		Feedback      string `json:"feedback"`
		Suggestions   string `json:"suggestions"`
		CorrectAnswer string `json:"correct_answer"`
	}

	// 清理响应内容，移除可能的markdown代码块标记
	cleanResponse := strings.TrimSpace(response)
	cleanResponse = strings.TrimPrefix(cleanResponse, "```json")
	cleanResponse = strings.TrimSuffix(cleanResponse, "```")
	cleanResponse = strings.TrimSpace(cleanResponse)

	err = json.Unmarshal([]byte(cleanResponse), &jsonResponse)
	if err != nil {
		// 如果JSON解析失败，使用传统方法解析
		score := ts.extractScore(response)
		feedback := response
		suggestions := ts.extractSuggestions(response)

		return &HomeworkResponse{
			Score:       score,
			Feedback:    feedback,
			Suggestions: suggestions,
			Success:     true,
		}, nil
	}

	// 使用JSON解析的结果
	suggestions := []string{}
	if jsonResponse.Suggestions != "" {
		// 将建议字符串按行分割
		suggestions = strings.Split(strings.TrimSpace(jsonResponse.Suggestions), "\n")
		// 清理空行
		var cleanSuggestions []string
		for _, s := range suggestions {
			if trimmed := strings.TrimSpace(s); trimmed != "" {
				cleanSuggestions = append(cleanSuggestions, trimmed)
			}
		}
		suggestions = cleanSuggestions
	}

	return &HomeworkResponse{
		Score:       jsonResponse.Score,
		Feedback:    jsonResponse.Feedback,
		Suggestions: suggestions,
		Success:     true,
	}, nil
}

// ProcessSubtitle 处理视频字幕
func (ts *TaskService) ProcessSubtitle(ctx context.Context, req SubtitleRequest) (*SubtitleResponse, error) {
	// 第一步：提取字幕（需要 ffmpeg）
	subtitlePath, err := ts.extractSubtitle(req.VideoPath)
	if err != nil {
		return &SubtitleResponse{
			Success: false,
			Message: fmt.Sprintf("字幕提取失败: %v", err),
		}, err
	}

	// 第二步：读取字幕内容
	content, err := os.ReadFile(subtitlePath)
	if err != nil {
		return &SubtitleResponse{
			Success: false,
			Message: fmt.Sprintf("读取字幕文件失败: %v", err),
		}, err
	}

	// 第三步：翻译字幕
	if req.TargetLang != req.SourceLang {
		translatedContent, err := ts.translateSubtitle(ctx, string(content), req.SourceLang, req.TargetLang)
		if err != nil {
			return &SubtitleResponse{
				Success: false,
				Message: fmt.Sprintf("字幕翻译失败: %v", err),
			}, err
		}
		content = []byte(translatedContent)
	}

	// 第四步：保存处理后的字幕
	outputPath := strings.Replace(subtitlePath, ".srt", "_processed.srt", 1)
	err = os.WriteFile(outputPath, content, 0o644)
	if err != nil {
		return &SubtitleResponse{
			Success: false,
			Message: fmt.Sprintf("保存字幕文件失败: %v", err),
		}, err
	}

	return &SubtitleResponse{
		SubtitlePath: outputPath,
		Content:      string(content),
		Success:      true,
	}, nil
}

// extractSubtitle 使用 ffmpeg 提取视频字幕
func (ts *TaskService) extractSubtitle(videoPath string) (string, error) {
	outputPath := strings.Replace(videoPath, filepath.Ext(videoPath), ".srt", 1)

	cmd := exec.Command("ffmpeg", "-i", videoPath, "-map", "0:s:0", "-c:s", "srt", outputPath, "-y")
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ffmpeg 提取字幕失败: %v", err)
	}

	return outputPath, nil
}

// translateSubtitle 翻译字幕内容
func (ts *TaskService) translateSubtitle(ctx context.Context, content, sourceLang, targetLang string) (string, error) {
	prompt := fmt.Sprintf(`请将以下%s字幕翻译为%s，保持时间戳格式不变：

%s

要求：
1. 保持SRT字幕格式
2. 保持时间戳不变
3. 翻译要准确自然
4. 保持字幕的分段结构`, sourceLang, targetLang, content)

	return ts.modelManager.ChatWithModel(ctx, "qwen2-7b-instruct", prompt, 4096)
}

// extractScore 从批改结果中提取分数
func (ts *TaskService) extractScore(response string) int {
	// 简单的分数提取逻辑，实际应用中可以使用正则表达式
	if strings.Contains(response, "100") || strings.Contains(response, "满分") {
		return 100
	} else if strings.Contains(response, "90") || strings.Contains(response, "优秀") {
		return 90
	} else if strings.Contains(response, "80") || strings.Contains(response, "良好") {
		return 80
	} else if strings.Contains(response, "70") || strings.Contains(response, "中等") {
		return 70
	} else if strings.Contains(response, "60") || strings.Contains(response, "及格") {
		return 60
	}
	return 75 // 默认分数
}

// extractSuggestions 从批改结果中提取建议
func (ts *TaskService) extractSuggestions(response string) []string {
	suggestions := []string{}
	lines := strings.Split(response, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "建议") || strings.Contains(line, "改进") {
			suggestions = append(suggestions, line)
		}
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "继续努力，保持学习热情")
	}

	return suggestions
}
