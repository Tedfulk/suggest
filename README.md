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

### Enhance Prompts

The `suggest` CLI provides two ways to enhance your coding-related prompts for better results:

1. Using the `-e` flag:
```bash
suggest -e "How do I use generics?"  # Enhances the prompt before processing
```

2. Using the enhance command:
```bash
suggest enhance "How do I use generics?"  # Same as above but as a separate command
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

| Command                                                                                                                           | Description                                       |
| --------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------- |
| `suggest template select translate --vars "source=English,target=French,text=Hello world" --system "You are a helpful assistant"` | Use the template with variables and system prompt |
| `suggest -s "Technical Writer" template select docs --vars "topic=API,format=markdown"`                                           | Use the template with variables and system prompt |
