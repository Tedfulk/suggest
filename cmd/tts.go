package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/tedfulk/suggest/internal/api"
	"github.com/tedfulk/suggest/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	ttsSpeed   int
	ttsVoice   string
	useGroqTTS bool
	useHumeTTS bool
)

// Additional color functions
var (
	red     = color.New(color.FgRed).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
)

var ttsCmd = &cobra.Command{
	Use:   "tts [text]",
	Short: "Convert text to speech using TTS services",
	Long: `Convert text to speech using various TTS services.
On macOS, uses the built-in 'say' command by default.
On Linux and other systems, uses Groq TTS API (requires Groq API key).

Example:
  suggest tts "Explain how Bitcoin mining works"
  suggest tts --speed 200 "What is the difference between proof of work and proof of stake?"  # Faster speech (macOS only)
  suggest tts --voice Fritz-PlayAI "How do smart contracts function on Ethereum?"
  suggest tts --use-groq --voice Mikail-PlayAI 'Why are transaction fees important in cryptocurrencies?'
  suggest tts --use-hume --voice "Booming American Narrator" "How do smart contracts function on Ethereum?"
  suggest tts --voice list
  echo "What is a blockchain fork?" | suggest tts`,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle voice listing
		if ttsVoice == "list" {
			listAvailableVoices()
			return
		}

		var message string

		// Check if data is being piped
		stat, _ := os.Stdin.Stat()
		isPiped := (stat.Mode() & os.ModeCharDevice) == 0

		if isPiped {
			// Read from stdin
			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Printf("%s: %v\n", red("Error reading from stdin"), err)
				return
			}
			message = strings.TrimSpace(string(bytes))
		} else if len(args) > 0 {
			// Use arguments
			message = strings.Join(args, " ")
		} else {
			fmt.Println(red("Please provide text to process via arguments or pipe content."))
			fmt.Printf("Example: %s\n", cyan("suggest tts \"Hello, world!\""))
			fmt.Printf("Example: %s\n", cyan("echo \"Hello, world!\" | suggest tts"))
			return
		}

		if strings.TrimSpace(message) == "" {
			fmt.Println(red("Error: No text provided to process"))
			return
		}

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("%s: %v\n", red("Error loading config"), err)
			return
		}

		// Get model and system prompt
		model := cfg.Model
		if modelFlag != "" {
			model = modelFlag
		}

		if actualModel, exists := cfg.ModelAliases[model]; exists {
			model = actualModel
		}

		systemPrompt := cfg.SystemPrompt
		if systemFlag != "" {
			var found bool
			for _, p := range cfg.SystemPrompts {
				if p.Title == systemFlag {
					systemPrompt = p.Content
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("%s: %s\n", red("System prompt not found"), yellow(systemFlag))
				return
			}
		}

		// Prepare messages for AI
		messages := []api.ChatMessage{}
		
		if systemPrompt != "" {
			messages = append(messages, api.ChatMessage{
				Role:    "system",
				Content: systemPrompt,
			})
		}
		
		messages = append(messages, api.ChatMessage{
			Role:    "user",
			Content: message,
		})

		req := &api.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: 0.7,
		}

		// Get response from AI
		var resp *api.ChatCompletionResponse
		var apiErr error

		provider := config.DetermineModelProvider(model, cfg)
		if provider == "" {
			fmt.Printf("%s: %s. Please use a Groq, OpenAI, Gemini, or Ollama model.\n", red("Model not supported"), yellow(model))
			return
		}

		switch provider {
		case "groq":
			if cfg.GroqAPIKey == "" {
				fmt.Printf("%s. Please set it in your config file.\n", red("Groq API key not set"))
				return
			}
			client := api.NewGroqClient(cfg.GroqAPIKey)
			resp, apiErr = client.CreateChatCompletion(req)
		
		case "openai":
			if cfg.OpenAIAPIKey == "" {
				fmt.Printf("%s. Please set it in your config file.\n", red("OpenAI API key not set"))
				return
			}
			client := api.NewOpenAIClient(cfg.OpenAIAPIKey)
			resp, apiErr = client.CreateChatCompletion(req)
		
		case "gemini":
			if cfg.GeminiAPIKey == "" {
				fmt.Printf("%s. Please set it in your config file.\n", red("Gemini API key not set"))
				return
			}
			client := api.NewGeminiClient(cfg.GeminiAPIKey)
			resp, apiErr = client.CreateChatCompletion(req)
		
		case "ollama":
			client := api.NewOllamaClient(cfg.OllamaHost)
			resp, apiErr = client.CreateChatCompletion(req)
		
		default:
			fmt.Printf("%s: %s. Please use a Groq, OpenAI, Gemini, or Ollama model.\n", red("Model not supported"), yellow(model))
			return
		}

		if apiErr != nil {
			fmt.Printf("%s: %v\n", red("Error getting AI response"), apiErr)
			return
		}

		if len(resp.Choices) == 0 {
			fmt.Println(red("Error: No response received from AI model"))
			return
		}

		// Get the AI response text
		aiResponse := resp.Choices[0].Message.Content
		
		// Clean up the response for speech (remove markdown, etc.)
		cleanResponse := cleanTextForSpeech(aiResponse)

		fmt.Println(green(cleanResponse))

		// Handle TTS based on platform or force flag
		if runtime.GOOS == "darwin" && !useGroqTTS && !useHumeTTS {
			// Use macOS say command (default on macOS)
			sayArgs := []string{}
			if ttsSpeed > 0 {
				sayArgs = append(sayArgs, "-r", strconv.Itoa(ttsSpeed))
			}
			sayArgs = append(sayArgs, cleanResponse)

			sayCmd := exec.Command("say", sayArgs...)
			sayCmd.Stdout = os.Stdout
			sayCmd.Stderr = os.Stderr

			err = sayCmd.Run()
			if err != nil {
				fmt.Printf("%s: %v\n", red("Error executing say command"), err)
				return
			}
		} else if useHumeTTS {
			// Use Hume TTS API
			if cfg.HumeAPIKey == "" {
				fmt.Printf("%s: %s\n", red("Error"), red("Hume API key required for TTS"))
				fmt.Printf("Please set your Hume API key: %s\n", cyan("suggest keys hume"))
				return
			}

			// Set default voice description if not specified
			if ttsVoice == "" {
				ttsVoice = "Booming American Narrator"
			}

			client := api.NewHumeClient(cfg.HumeAPIKey)
			audioData, err := client.CreateTTS(cleanResponse, ttsVoice)
			if err != nil {
				fmt.Printf("%s: %v\n", red("Error generating speech"), err)
				return
			}

			// Play the audio using a system command
			err = playAudio(audioData)
			if err != nil {
				fmt.Printf("%s: %v\n", red("Error playing audio"), err)
				return
			}
		} else {
			// Use Groq TTS API (non-macOS systems or when --use-groq is set)
			if cfg.GroqAPIKey == "" {
				fmt.Printf("%s: %s\n", red("Error"), red("Groq API key required for TTS"))
				fmt.Printf("Please set your Groq API key: %s\n", cyan("suggest keys groq"))
				return
			}

			// Set default voice if not specified
			if ttsVoice == "" {
				ttsVoice = "Fritz-PlayAI" // Default Groq voice
			}

			client := api.NewGroqClient(cfg.GroqAPIKey)
			audioData, err := client.CreateTTS(cleanResponse, ttsVoice)
			if err != nil {
				fmt.Printf("%s: %v\n", red("Error generating speech"), err)
				return
			}

			// Play the audio using a system command
			err = playAudio(audioData)
			if err != nil {
				fmt.Printf("%s: %v\n", red("Error playing audio"), err)
				return
			}
		}
	},
}

// listAvailableVoices lists all available TTS voices
func listAvailableVoices() {
	// Groq TTS voices
	groqVoices := map[string]string{
		"Arista-PlayAI":   "English",
		"Atlas-PlayAI":    "English",
		"Basil-PlayAI":    "English",
		"Briggs-PlayAI":   "English",
		"Calum-PlayAI":    "English",
		"Celeste-PlayAI":  "English",
		"Cheyenne-PlayAI": "English",
		"Chip-PlayAI":     "English",
		"Cillian-PlayAI":  "English",
		"Deedee-PlayAI":   "English",
		"Fritz-PlayAI":    "English",
		"Gail-PlayAI":     "English",
		"Indigo-PlayAI":   "English",
		"Mamaw-PlayAI":    "English",
		"Mason-PlayAI":    "English",
		"Mikail-PlayAI":   "English",
		"Mitch-PlayAI":    "English",
		"Quinn-PlayAI":    "English",
		"Thunder-PlayAI":  "English",
	}

	// Hume TTS voice examples
	humeVoices := map[string]string{
		"Booming American Narrator":                                     "Dramatic, authoritative",
		"Middle-aged masculine voice with a clear, rhythmic Scots lilt": "Academic, warm tone",
		"Young female voice with bright, energetic delivery":             "Energetic, friendly",
		"Deep male voice with authoritative, professional tone":          "Professional, authoritative",
		"Soft-spoken female voice with gentle, caring demeanor":         "Gentle, caring",
		"Elderly male voice with wisdom and experience":                 "Wise, experienced",
	}

	fmt.Println(cyan("Available TTS Voices:"))
	fmt.Println(cyan("====================="))
	
	fmt.Printf("\n%s:\n", blue("Groq TTS Voices"))
	fmt.Println(blue("---------------"))
	for voice, style := range groqVoices {
		fmt.Printf("  %s -- %s\n", yellow(voice), style)
	}
	
	fmt.Printf("\n%s:\n", magenta("Hume TTS Voice Examples"))
	fmt.Println(magenta("------------------------"))
	for voice, style := range humeVoices {
		fmt.Printf("  %s -- %s\n", yellow(voice), style)
	}
	
	fmt.Printf("\n%s:\n", green("Usage"))
	fmt.Printf("  %s\n", cyan("suggest tts --voice <voice_name> \"Your text here\""))
	fmt.Printf("  %s\n", cyan("suggest tts --use-groq --voice Fritz-PlayAI \"Hello, world!\""))
	fmt.Printf("  %s\n", cyan("suggest tts --use-groq --voice Mikail-PlayAI \"Hello, world!\""))
	fmt.Printf("  %s\n", cyan("suggest tts --use-hume --voice \"Booming American Narrator\" \"Hello, world!\""))
}

// playAudio plays audio data using system commands
func playAudio(audioData []byte) error {
	// Create a temporary file with .wav extension
	tmpFile, err := os.CreateTemp("", "suggest-tts-*.wav")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write audio data to file
	_, err = tmpFile.Write(audioData)
	if err != nil {
		return fmt.Errorf("error writing audio data: %w", err)
	}
	tmpFile.Close()

	// Play audio using system command
	var playCmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		// Try different audio players on macOS
		players := []string{"ffplay", "afplay", "mpv", "cvlc"}
		for _, player := range players {
			if _, err := exec.LookPath(player); err == nil {
				switch player {
				case "ffplay":
					playCmd = exec.Command("ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", tmpFile.Name())
				case "afplay":
					playCmd = exec.Command("afplay", "-q", "1", tmpFile.Name())
				case "mpv":
					playCmd = exec.Command("mpv", "--no-video", "--really-quiet", tmpFile.Name())
				case "cvlc":
					playCmd = exec.Command("cvlc", "--play-and-exit", "--quiet", tmpFile.Name())
				}
				break
			}
		}
		if playCmd == nil {
			// Fallback to afplay even if not found in PATH
			playCmd = exec.Command("afplay", "-q", "1", tmpFile.Name())
		}
	case "linux":
		// Try different audio players on Linux
		players := []string{"aplay", "paplay", "ffplay", "mpv", "cvlc"}
		for _, player := range players {
			if _, err := exec.LookPath(player); err == nil {
				switch player {
				case "aplay":
					playCmd = exec.Command("aplay", tmpFile.Name())
				case "paplay":
					playCmd = exec.Command("paplay", tmpFile.Name())
				case "ffplay":
					playCmd = exec.Command("ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", tmpFile.Name())
				case "mpv":
					playCmd = exec.Command("mpv", "--no-video", "--really-quiet", tmpFile.Name())
				case "cvlc":
					playCmd = exec.Command("cvlc", "--play-and-exit", "--quiet", tmpFile.Name())
				}
				break
			}
		}
		if playCmd == nil {
			return fmt.Errorf("no audio player found. Please install one of: aplay, paplay, ffplay, mpv, or vlc")
		}
	default:
		return fmt.Errorf("audio playback not supported on %s", runtime.GOOS)
	}

	playCmd.Stdout = os.Stdout
	playCmd.Stderr = os.Stderr
	return playCmd.Run()
}

// cleanTextForSpeech removes markdown formatting and other elements that don't work well with speech
func cleanTextForSpeech(text string) string {
	// Remove markdown code blocks
	text = strings.ReplaceAll(text, "```", "")
	
	// Remove markdown headers
	lines := strings.Split(text, "\n")
	var cleanedLines []string
	
	for _, line := range lines {
		// Skip lines that are just markdown formatting
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "**") && strings.HasSuffix(strings.TrimSpace(line), "**") {
			// Convert bold text to regular text
			line = strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(line), "**"), "**")
		}
		if strings.HasPrefix(strings.TrimSpace(line), "*") && strings.HasSuffix(strings.TrimSpace(line), "*") {
			// Convert italic text to regular text
			line = strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(line), "*"), "*")
		}
		cleanedLines = append(cleanedLines, line)
	}
	
	// Join lines and clean up extra whitespace
	result := strings.Join(cleanedLines, " ")
	result = strings.ReplaceAll(result, "  ", " ")
	result = strings.TrimSpace(result)
	
	return result
}

func init() {
	ttsCmd.Flags().IntVarP(&ttsSpeed, "speed", "r", 0, "Speech rate (words per minute, macOS only)")
	ttsCmd.Flags().StringVarP(&ttsVoice, "voice", "v", "", "Voice name for TTS (non-macOS only, default: Fritz-PlayAI)")
	ttsCmd.Flags().BoolVar(&useGroqTTS, "use-groq", false, "Force Groq TTS (any platform)")
	ttsCmd.Flags().BoolVar(&useHumeTTS, "use-hume", false, "Force Hume TTS (any platform)")
	rootCmd.AddCommand(ttsCmd)
} 