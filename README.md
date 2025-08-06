# suggest

A powerful command-line tool for interacting with AI models through various providers. Streamline your AI interactions with customizable templates, system prompts, and model management.

## Installation

### Using Homebrew (recommended)

```bash
brew tap tedfulk/suggest
brew install suggest
```

### Using Go Install

```bash
go install github.com/tedfulk/suggest@latest
```

## Usage

### Basic Usage

Provide your prompt directly as arguments:

```bash
suggest "Tell me a joke about Go"
suggest -m gpt-4 "Explain the difference between channels and mutexes"
```

### Using Piped Input

You can pipe content (like file contents) directly into `suggest` to provide context. The arguments will then act as the specific prompt or question about the piped context.

```bash
# Pipe a file and ask a question about it
cat main.go | suggest "Explain what this Go code does"

# Pipe content and use a specific model and system prompt
cat report.txt | suggest -m gemini-1.5-pro -s "Summarizer" "Provide a 5-bullet summary"

# Pipe content without arguments (AI will try to infer the task)
cat README.md | suggest
```

In Fish shell, you can combine multiple commands' output into the pipe using `begin...end`:

```fish
begin
  echo 'My name is Teddy.'
  echo '-- Context Separator --'
  cat file.txt
end | suggest "Summarize the file content after the separator. Also, what's my name?"
```

### Interactive Chat Mode

Start an interactive chat session with your preferred AI model:

```bash
suggest chat                    # Start chat with default model
suggest chat -m gpt-4          # Chat with a specific model
suggest chat -s "Coder"        # Chat with a specific system prompt
```

To exit the chat session, type 'bye', 'stop', 'end', or press Ctrl+C.

### Set Your Chat Username

```bash
suggest username  # Set your username for chat sessions
```

### Text-to-Speech (TTS)

Convert text to speech using various TTS services. On macOS, uses the built-in `say` command by default. On Linux and other systems, uses Groq TTS API (requires Groq API key). Also supports Hume TTS for high-quality voice synthesis.

```bash
suggest tts "Explain how Bitcoin mining works"
suggest tts --speed 200 "What is the difference between proof of work and proof of stake?"  # Faster speech (macOS only)
suggest tts --voice Fritz-PlayAI "How do smart contracts function on Ethereum?"
suggest tts --use-groq --voice Mikail-PlayAI 'Why are transaction fees important in cryptocurrencies?'
suggest tts --use-hume --voice "Booming American Narrator" "How do smart contracts function on Ethereum?"
suggest tts --voice list                       # List available voices (non-macOS)
echo "What is a blockchain fork?" | suggest tts
```

**Note:** 
- **macOS**: Uses built-in `say` command (no setup required)
- **Linux/Other**: Requires Groq API key
  - Set Groq API key: `suggest keys groq`
  - Available Groq voices: Fritz-PlayAI, Celeste-PlayAI, Atlas-PlayAI, and 16 others
  - List voices: `suggest tts --voice list`
- **Testing**: Use `--use-groq` to test Groq TTS on any platform
- **Hume TTS**: High-quality voice synthesis with customizable voice descriptions
  - Set Hume API key: `suggest keys hume`
  - Use `--use-hume` to force Hume TTS on any platform
  - Voice descriptions allow for detailed voice customization

### Enhance Prompts

The `suggest` CLI provides two ways to enhance your coding-related prompts for better results:

1. Using the `-e` flag:
```bash
suggest -e "How do I use generics in Typescript?"  # Enhances the prompt before processing
```

2. Using the enhance command:
```bash
suggest enhance "How do I use generics in Typescript?"  # Same as above but as a separate command
```

Both methods will:
1. Use Groq's llama-3.3-70b-versatile model to enhance your prompt with more specificity and structure
2. Show you the enhanced version
3. Process the enhanced prompt with your default model

Example:
```bash
suggest enhance "What are design patterns?"
# Will enhance the prompt to be more specific and detailed before processing
```

Note: The enhance feature requires a Groq API key to be configured.

### Generate default config

```bash
suggest generate
```

View current configuration

```bash
cat ~/.suggest/config.yaml
```

### Use different models for different tasks

```bash
suggest "What is the difference between public and private in typescript" # This will use the default model you have set in config
suggest -m qwen-qwq-32b "Explain go routines" # This will use the specified modle you have passed in
```

### Ollama Integration

`suggest` supports local AI models through Ollama. To use Ollama models:

1. Install Ollama from https://ollama.ai
2. Pull your desired models:
   ```bash
   ollama pull codellama    # For coding tasks
   ollama pull llama3       # General purpose
   ```
3. Configure Ollama host (optional, defaults to http://localhost:11434):
   ```bash
   suggest keys ollama
   ```
4. Use Ollama models:
   ```bash
   suggest model            # Select model interactively
   # or
   suggest -m codellama "Write a Python function"
   ```

Ollama runs locally on your machine, so no API key is required. You can use any model that you've pulled with `ollama pull`.

### API Key

| Command                           | Description                          |
| --------------------------------- | ------------------------------------ |
| `suggest keys openai sk-proj-...` | Set OpenAI API key directly          |
| `suggest keys groq gr-...`        | Set Groq API key directly            |
| `suggest keys gemini gl-...`      | Set Gemini API key directly          |
| `suggest keys openai`             | Set OpenAI API key interactively     |
| `suggest keys groq`               | Set Groq API key interactively       |
| `suggest keys gemini`             | Set Gemini API key interactively     |
| `suggest keys `                   | Select and set API key interactively |

### Model Management

| Command                                        | Description                                |
| ---------------------------------------------- | ------------------------------------------ |
| `suggest models`                               | List all available models                  |
| `suggest models --update`                      | Update model list                          |
| `suggest model`                                | Interactively select a model               |
| `suggest alias add g1.5 gemini-1.5-pro`        | Create model alias for gemini-1.5-pro      |
| `suggest alias list`                           | List all model aliases                     |
| `suggest alias remove g1.5`                    | Remove a model alias                       |

### System Prompts

| Command                                                        | Description                        |
| -------------------------------------------------------------- | ---------------------------------- |
| `suggest system add`                                           | Add a system prompt (interactive)  |
| `suggest system add "coder" "You are an expert programmer..."` | Add a system prompt (direct)       |
| `suggest system list`                                          | List system prompts                |
| `suggest system select`                                        | Select system prompt interactively |
| `suggest system select "coder"`                                | Select system prompt directly      |
| `suggest system remove "coder"`                                | Remove a system prompt             |

### Enhance Feature

| Command                                                        | Description                                                |
| -------------------------------------------------------------- | ---------------------------------------------------------- |
| `suggest -e "Write a Python function"`                         | Enhance prompt with additional context and clarification   |
| `suggest --enhance "Explain quantum computing"`                | Expand prompt to be more detailed and specific            |
| `suggest -e -m gpt-4 "Create a web scraper"`                  | Use enhancement with a specific model                      |
| `suggest -e -s "coder" "Implement binary search"`             | Combine enhancement with system prompt                     |

This feature requires a Groq API key to be configured:

### Templates

| Command                                                                  | Description                  |
| ------------------------------------------------------------------------ | ---------------------------- |
| `suggest template add`                                                   | Add a template (interactive) |
| `suggest template add "code" "Write a [language] function that [task]"`  | Add a template (direct)      |
| `suggest template list`                                                  | List templates               |
| `suggest template select`                                                | Use template interactively   |
| `suggest template select code --vars "language=Python,task=sort a list"` | Use template with variables  |
| `suggest template remove code`                                           | Remove template              |

### Advanced Usage

#### Template Variables

Templates support variable substitution using square brackets:

| Command                                                                                    | Description                      |
| ------------------------------------------------------------------------------------------ | -------------------------------- |
| `suggest template add "translate" "Translate this text from [source] to [target]: [text]"` | Create a template with variables |
| `suggest template select translate --vars "source=English,target=French,text=Hello world"` | Use the template with variables  |

#### System Prompt Chaining

You can combine system prompts with templates:

| Command                                                                                                                           |
| --------------------------------------------------------------------------------------------------------------------------------- |
| `suggest template select translate --vars "source=English,target=French,text=Hello world" --system "You are a helpful assistant"` |
| `suggest -s "Technical Writer" template select docs --vars "topic=API,format=markdown"`                                           |
