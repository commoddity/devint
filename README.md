<div align="center">
<h1>`devint` - Developer Interface</h1>

</div>
<br/>

## Table of Contents <!-- omit in toc -->

- [Overview](#overview)
- [Providers](#providers)
- [Thaura](#thaura)
- [DeepSeek](#deepseek)
- [OpenRouter 🌐](#openrouter-)
- [Interactive Chat](#interactive-chat)
  - [One-Shot Mode](#one-shot-mode)
  - [Interactive Mode](#interactive-mode)
- [Usage](#usage)
- [Configuration](#configuration)


## Overview

The Developer Interface (DI) is a command-line tool designed to streamline developer workflows. DI helps developers quickly perform routine operations and maintain consistency across projects.

DI integrates with multiple Language Model (LLM) providers to power intelligent features like automated PR creation, code summarization, and more. The tool provides a unified interface for interacting with various LLM providers, making it easy to switch between different models and services.

Key features include:
- 🤖 **Multi-Provider LLM Support**: Seamlessly switch between Thaura, DeepSeek, and OpenRouter
- 🔧 **Git Integration**: Automated PR creation and summarization powered by AI
- ⚙️ **Flexible Configuration**: Easy-to-use YAML-based configuration system
- 🎯 **Developer-Focused**: Built for developers who want to automate routine tasks

## Providers

DI supports multiple LLM providers, each offering unique capabilities and models. You can configure one or more providers and switch between them as needed.

| Provider         | Website                                | Description                                         | Models                                                                      |
| ---------------- | -------------------------------------- | --------------------------------------------------- | --------------------------------------------------------------------------- |
| 🇵🇸 **Thaura**     | [thaura.ai](https://thaura.ai)         | AI platform supporting Palestinian liberation       | `thaura`                                                                    |
| 🤖 **DeepSeek**   | [deepseek.com](https://deepseek.com)   | High-performance AI models for coding and reasoning | `deepseek-chat`, `deepseek-reasoner`                                        |
| 🌐 **OpenRouter** | [openrouter.ai](https://openrouter.ai) | Unified API for accessing multiple AI models        | Any model string (see [openrouter.ai/models](https://openrouter.ai/models)) |

## Thaura 

<div align="center">
  <a href="https://thaura.ai">
    <img  width="400" src=".github/img/thaura.png" alt="Thaura Logo" />
  </a>
  <br/>
  <p> <a href="https://thaura.ai">Thaura</a> is an AI platform that combines technical excellence with ethical principles, designed to support Palestinian liberation and mission-aligned technology development.</p>
</div>
<br/>

**Ethical Principles:**
- **Mission-Aligned Technology**: Supports projects and organizations working towards Palestinian liberation
- **Transparent Operations**: Clear documentation and open communication about platform capabilities
- **Ethical AI Development**: Committed to responsible AI practices that prioritize fairness and accountability
- **Community-Driven**: Built in collaboration with the Tech for Palestine community

For API documentation, see [thaura.ai/api-platform](https://thaura.ai/api-platform)


> ## 🇵🇸 Tech for Palestine <!-- omit in toc -->
> [Tech for Palestine](https://techforpalestine.org/) (T4P) is a coalition of thousands of founders, engineers, product marketers, investors, and other professionals working in support of Palestinian liberation.
>
>**What is Tech for Palestine?**
>
>Tech for Palestine is first and foremost an incubator for advocacy projects. They rally thousands of volunteers from across the tech world — founders, engineers, marketers, investors, and more — all committed to Palestinian liberation.
>
>The T4P Incubator helps pro-Palestine advocates build, grow, and scale their work towards a Free Palestine. They support projects — whether collections of individuals, registered non-profits, or even companies — whose mission helps Palestine, especially advocacy groups building technical products or in the tech space.
>
>The Incubator is free and provides:
>- 👥 **Volunteers** - Access to skilled professionals
>- 📢 **Marketing Support** - Help spreading your message
>- 🎓 **Mentorship** - Guidance from experienced professionals
>- 🔗 **Connections** - Links to the broader Palestinian advocacy ecosystem
>
>
>**Get Involved:**
>- Volunteer your skills
>- Join their Discord
>- Start a project of your own
>- Be a mentor
>- Hire Palestinians
>
>Learn more at [techforpalestine.org](https://techforpalestine.org/)

## DeepSeek

[DeepSeek](https://epseek.com) provides powerful AI models optimized for coding and reasoning tasks. DeepSeek models are known for their excellent performance in technical contexts.

**Available Models:**
- `deepseek-chat` - General-purpose chat model
- `deepseek-reasoner` - Advanced reasoning model

## OpenRouter 🌐

[OpenRouter](https://openrouter.ai) is a unified API that provides access to multiple AI models from various providers. It offers flexibility and choice, allowing you to use different models through a single interface.

**Available Models:**
OpenRouter accepts any model string. You can use any model available on OpenRouter's platform. 

See [openrouter.ai/models](https://openrouter.ai/models) for the full list of available models.

## Interactive Chat

The `devint chat` command provides a powerful interface for interacting with your configured LLM provider. It supports two modes of operation:

### One-Shot Mode

Send a single prompt and receive a response:

```bash
# Using the default provider
devint chat "tell me about winnie the pooh?"

# Override the provider
devint chat -p thaura "what is the capital of France?"

# Override the model
devint chat -m deepseek-reasoner "solve this complex problem..."
```

### Interactive Mode

Start an interactive chat session that preserves conversation context throughout the session:

```bash
devint chat
```

In interactive mode, you can:
- 💬 Have natural conversations with the AI that remembers previous messages
- ⌨️ Use arrow keys to navigate and edit your input
- 📜 Access command history with up/down arrows
- 🎨 See beautifully formatted responses with proper markdown rendering (headings, bold text, lists)
- 🛑 Exit anytime with `Ctrl+C`

The interactive mode maintains full conversation context, allowing the AI to reference previous questions and answers in the same session.

<div align="center">
  <img src=".github/img/demo.png" alt="devint chat interactive mode" width="900"/>
</div>


## Usage

The Developer Interface (DI) enables streamlined development workflows by providing a unified command-line interface to manage configuration settings, execute Git operations, and interact with AI models. Below are tables of available commands and their flags:

### devint <!-- omit in toc -->

| Flag         | Type | Required | Description                          |
| ------------ | ---- | -------- | ------------------------------------ |
| -t, --toggle | bool | ❌        | Toggle verbose mode or other options |
| -h, --help   | bool | ❌        | Show help for devint                 |

### devint chat <!-- omit in toc -->

| Flag                     | Type   | Required | Description                                                        |
| ------------------------ | ------ | -------- | ------------------------------------------------------------------ |
| [prompt]                 | string | ❌        | Prompt to send to the LLM (if omitted, starts interactive mode)    |
| --provider-override (-p) | string | ❌        | LLM provider override. Sets the LLM provider only for this request |
| --model-override (-m)    | string | ❌        | LLM model override. Sets the LLM model only for this request       |

### devint git createpr <!-- omit in toc -->

| Flag                     | Type   | Required | Description                                                                         |
| ------------------------ | ------ | -------- | ----------------------------------------------------------------------------------- |
| --pr-title (-t)          | string | ✅        | PR title. Will open a draft PR if the string contains [DRAFT] or [WIP]              |
| --target-branch (-b)     | string | ❌        | Target branch (default "main")                                                      |
| --issue (-i)             | int    | ❌        | Issue number                                                                        |
| --dummy (-d)             | bool   | ❌        | Dummy mode. Will print summary to console and clipboard but not open a PR on GitHub |
| --provider-override (-p) | string | ❌        | LLM provider override. Sets the LLM provider only for this request                  |
| --model-override (-m)    | string | ❌        | LLM model override. Sets the LLM model only for this request                        |

### devint git summarizepr <!-- omit in toc -->

| Flag              | Type   | Required | Description                                                                    |
| ----------------- | ------ | -------- | ------------------------------------------------------------------------------ |
| --pr-number (-p)  | int    | ✅        | Pull request number to summarize                                               |
| --repo-owner (-r) | string | ❌        | GitHub repository owner override. If set, overrides the repo_owner from config |

This command fetches a pull request by number, retrieves its diff, and saves a formatted markdown summary to the directory specified by `pr_summary_output_dir` in your config file. The summary includes PR metadata (title, number, status, author, URL), description, and formatted diff.

**Note:** The `pr_summary_output_dir` must be set in your `git_config` (via `devint config`) for this command to work.

### devint config <!-- omit in toc -->  

| Flag          | Type   | Required | Description                                                                  |
| ------------- | ------ | -------- | ---------------------------------------------------------------------------- |
| --show (-s)   | bool   | ❌        | Show the configuration                                                       |
| --editor (-e) | string | ❌        | Edit the configuration in the given text editor, for example `nano` or `vim` |

To run the interactive editor, run without any flags, ie. `devint config`.

## Configuration

The configuration is done through a config YAML file located at `~/.devint.config.yaml`.

An example configuration file can be found at `config/examples/.config.example.yaml`.

Configuration can be updated using the interactive command `devint config`.

```yaml
# yaml-language-server: $schema=../config.schema.yaml
git_config:
  personal_access_token: github_pat_1a2b3c4d5e6f7g8h9i0j1a2b3c4d5e6f7g8h9i0j
  repo_owner: your-github-org-or-username
llm_config:
  default_llm_provider: deepseek
  llm_providers:
    deepseek:
      api_key: "your-deepseek-api-key"
      client_model: "deepseek-chat"
    openrouter:
      api_key: "your-openrouter-api-key"
      client_model: "deepseek/deepseek-chat"
    thaura:
      api_key: "your-thaura-api-key"
      client_model: "thaura"
```
