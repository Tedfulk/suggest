# suggest

A powerful command-line tool for interacting with AI models through various providers. Streamline your AI interactions with customizable templates, system prompts, and model management.

## Installation

```bash
go install github.com/tedfulk/suggest@latest
```

## Usage

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
