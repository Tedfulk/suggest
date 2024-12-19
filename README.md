# suggest

A powerful command-line tool for interacting with AI models through various providers. Streamline your AI interactions with customizable templates, system prompts, and model management.

## Installation

```bash
go install github.com/tedfulk/suggest@latest
```

## Usage

### Generate default config

```bash
suggest generate-config
```

View current configuration

```bash
cat ~/.suggest/config.yaml
```

### Use different models for different tasks

```bash
suggest "Complex reasoning task"
suggest -m mixtral "Code generation task"
```

### API Key

| Command | Description |
|---------|-------------|
| `suggest keys openai sk-proj-...` | Set OpenAI API key directly |
| `suggest keys groq gr-...` | Set Groq API key directly |
| `suggest keys gemini gl-...` | Set Gemini API key directly |
| `suggest keys openai` | Set OpenAI API key interactively |
| `suggest keys groq` | Set Groq API key interactively |
| `suggest keys gemini` | Set Gemini API key interactively |

### Model Management

| Command | Description |
|---------|-------------|
| `suggest models` | List all available models |
| `suggest models --update` | Update model list |
| `suggest model` | Interactively select a model |
| `suggest alias add gpt4 gpt-4-turbo-preview` | Create model alias for gpt-4-turbo-preview |
| `suggest alias add mixtral mixtral-8x7b-32768` | Create model alias for mixtral-8x7b-32768 |
| `suggest alias list` | List all model aliases |
| `suggest alias remove gpt4` | Remove a model alias |

### System Prompts

| Command | Description |
|---------|-------------|
| `suggest system add` | Add a system prompt (interactive) |
| `suggest system add "coder" "You are an expert programmer..."` | Add a system prompt (direct) |
| `suggest system list` | List system prompts |
| `suggest system select` | Select system prompt interactively |
| `suggest system select "coder"` | Select system prompt directly |
| `suggest system remove "coder"` | Remove a system prompt |

### Templates

| Command | Description |
|---------|-------------|
| `suggest template add` | Add a template (interactive) |
| `suggest template add "code" "Write a [language] function that [task]"` | Add a template (direct) |
| `suggest template list` | List templates |
| `suggest template select` | Use template interactively |
| `suggest template select code --vars "language=Python,task=sort a list"` | Use template with variables |
| `suggest template remove code` | Remove template |

### Advanced Usage

#### Template Variables

Templates support variable substitution using square brackets:

| Command | Description |
|---------|-------------|
| `suggest template add "translate" "Translate this text from [source] to [target]: [text]"` | Create a template with variables |
| `suggest template select translate --vars "source=English,target=French,text=Hello world"` | Use the template with variables |

#### System Prompt Chaining

You can combine system prompts with templates:

| Command | Description |
|---------|-------------|
| `suggest template select translate --vars "source=English,target=French,text=Hello world" --system "You are a helpful assistant"` | Use the template with variables and system prompt |
| `suggest -s "Technical Writer" template select docs --vars "topic=API,format=markdown"` | Use the template with variables and system prompt |

