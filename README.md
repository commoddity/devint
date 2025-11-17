<table align="center">
<tr>
<td align="center">

<pre>
╔═════════════════════════════════════════════════════╗
║                                                     ║
║   ██████╗ ███████╗██╗   ██╗██╗███╗   ██╗████████╗   ║
║   ██╔══██╗██╔════╝██║   ██║██║████╗  ██║╚══██╔══╝   ║
║   ██║  ██║█████╗  ██║   ██║██║██╔██╗ ██║   ██║      ║
║   ██║  ██║██╔══╝  ╚██╗ ██╔╝██║██║╚██╗██║   ██║      ║
║   ██████╔╝███████╗ ╚████╔╝ ██║██║ ╚████║   ██║      ║
║   ╚═════╝ ╚══════╝  ╚═══╝  ╚═╝╚═╝  ╚═══╝   ╚═╝      ║
║                                                     ║
╚═════════════════════════════════════════════════════╝

💻 Developer Interface

Developer toolkit powered by multi-provider LLM integration.
</pre>

</td>
</tr>
</table>
<br/>

<!-- GIF: devint chat introduction -->
<div align="center">
  <img src=".github/img/chat-introduction.gif" alt="devint chat introduction" width="1200"/>
  <p><i>Introduction to the Developer Interface</i></p>
</div>

## Table of Contents <!-- omit in toc -->

- [🚀 Quickstart](#-quickstart)
- [📖 Overview](#-overview)
- [⚗️ Providers](#️-providers)
  - [🇵🇸 Thaura](#-thaura)
  - [🐋 DeepSeek](#-deepseek)
  - [🌍 OpenRouter](#-openrouter)
- [💬 Interactive Chat](#-interactive-chat)
  - [⚡ One-Shot Mode](#-one-shot-mode)
  - [💭 Interactive Mode](#-interactive-mode)
- [🔧 Git Commands](#-git-commands)
- [⚙️ Configuration](#️-configuration)
- [📝 TODOs](#-todos)

# 🚀 Quickstart

## Dependencies <!-- omit in toc -->

- [Go](https://go.dev/dl/) (version 1.21 or later)

## Installation <!-- omit in toc -->

Install `devint` using Go's install command:

```bash
go install github.com/commoddity/devint@latest
```

To verify the installation:
```bash
devint --version
```

## First-Time Setup <!-- omit in toc -->

When you run `devint` for the first time, it will automatically launch an interactive setup prompt that guides you through configuration.

<details>
<summary><b>Click to see the values the setup wizard will ask you to configure</b></summary>
<br/>

1. **Git Configuration** (required):
   - GitHub Personal Access Token (PAT) with at least `write:repo` scope
   - GitHub repository owner (organization or username)
   - Optional: PR summary output directory
   - Optional: Company name to filter from diffs

2. **LLM Provider Configuration** (required):
   - Default LLM provider (choose from DeepSeek, OpenRouter, or Thaura)
   - API key for your chosen provider
   - Default model for your provider
   - Optionally configure additional providers
</details>
<br/>

Simply run any `devint` command to trigger the first-time setup:

```bash
devint chat
```

The wizard will guide you through each step with clear prompts and helpful descriptions.

## Running Your First Command <!-- omit in toc -->

Once setup is complete, try a one-shot query:

```bash
devint chat "Can you explain Manifest Destiny in simple terms?"
```

That's it! You're ready to use the Developer Interface. 🎉

# 📖 Overview

The Developer Interface (`devint`) is a command-line tool designed to streamline developer workflows. `devint` helps developers quickly perform routine operations and maintain consistency across projects.

`devint` integrates with multiple Language Model (LLM) providers to power intelligent features like automated PR creation, code summarization, and more. The tool provides a unified interface for interacting with various LLM providers, making it easy to switch between different models and services.

Key features include:
- 📡 **Multi-Provider LLM Support**: Seamlessly switch between Thaura, DeepSeek, and OpenRouter
- 🔧 **Git Automation**: Automated git tasks for streamlining your workflow
- ⚙️ **Flexible Configuration**: Easy-to-use YAML-based configuration system
- 🎯 **Developer-Focused**: Built for developers who want to automate routine tasks
- 💬 **Live Chat**: Interact with AI models in real-time from your terminal

# ⚗️ Providers

`devint` supports multiple LLM providers. You can configure one or more providers and switch between them as needed.

| Provider              | Website                                | Description                                         | Models                                                   |
| --------------------- | -------------------------------------- | --------------------------------------------------- | -------------------------------------------------------- |
| 🇵🇸&nbsp;**Thaura**     | [thaura.ai](https://thaura.ai)         | Ethical AI platform from Tech for Palestine         | `thaura`                                                 |
| 🐋&nbsp;**DeepSeek**   | [deepseek.com](https://deepseek.com)   | High-performance AI models for coding and reasoning | `deepseek-chat`, `deepseek-reasoner`                     |
| 🌍&nbsp;**OpenRouter** | [openrouter.ai](https://openrouter.ai) | Unified API for accessing AI models                 | See [openrouter.ai/models](https://openrouter.ai/models) |

## 🇵🇸 Thaura 

<div align="center">
  <a href="https://thaura.ai">
    <img  width="400" src=".github/img/thaura.png" alt="Thaura Logo" />
  </a>
  <br/>
  <p> <a href="https://thaura.ai">Thaura</a> is an AI platform that combines technical excellence with ethical principles, <br/>designed to support Palestinian liberation and mission-aligned technology development.</p>
</div>
<br/>

**Ethical Principles:**
- **Mission-Aligned Technology**: Supports projects and organizations working towards Palestinian liberation
- **Transparent Operations**: Clear documentation and open communication about platform capabilities
- **Ethical AI Development**: Committed to responsible AI practices that prioritize fairness and accountability
- **Community-Driven**: Built in collaboration with the Tech for Palestine community

For API documentation, see [thaura.ai/api-platform](https://thaura.ai/api-platform)


> ## 🇵🇸 Tech for Palestine <!-- omit in toc -->
> [Tech for Palestine](https://techforpalestine.org/) (T4P) is a coalition of founders, engineers, product marketers, investors, and other professionals working in support of Palestinian liberation.
>
>**What is Tech for Palestine?**
>
>Tech for Palestine is first and foremost an incubator for advocacy projects. They rally volunteers from across the tech world — founders, engineers, marketers, investors, and more — all committed to Palestinian liberation.
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

## 🐋 DeepSeek

<div align="center">
  <a href="https://deepseek.com">
    <img  width="400" src=".github/img/deepseek.png" alt="DeepSeek Logo" />
  </a>
</div>
<br/>

[DeepSeek](https://epseek.com) provides powerful AI models optimized for coding and reasoning tasks. DeepSeek models are known for their excellent performance in technical contexts.

**Available Models:**
- `deepseek-chat` - General-purpose chat model
- `deepseek-reasoner` - Advanced reasoning model

## 🌍 OpenRouter

<div align="center">
  <a href="https://openrouter.ai">
    <img  width="400" src=".github/img/openrouter.png" alt="OpenRouter Logo" />
  </a>
</div>
<br/>

[OpenRouter](https://openrouter.ai) is a unified API that provides access to multiple AI models from various providers. It offers flexibility and choice, allowing you to use different models through a single interface.

**Available Models:**
OpenRouter accepts any model string. You can use any model available on OpenRouter's platform. 

See [openrouter.ai/models](https://openrouter.ai/models) for the full list of available models.

# 💬 Interactive Chat

The `devint chat` command supports two modes of operation:

## ⚡ One-Shot Mode

<!-- GIF: devint chat one-shot mode -->
<div align="center">
  <img src=".github/img/chat-oneshot.gif" alt="devint chat one-shot mode" width="1200"/>
  <p><i>One-shot mode with streaming response</i></p>
</div>

Send a single prompt and receive a response directly in your terminal.

```bash
# Using the default provider
devint chat "Who is Captain Ibrahim Traore?"

# Override the provider
devint chat -p thaura "Tell me about the history of Brazilian music."

# Override the model
devint chat -p deepseek -m deepseek-reasoner "Help me build a complex system of philosophical development based on Marxist theory."
```

## 💭 Interactive Mode

<!-- GIF: devint chat interactive mode -->
<div align="center">
  <img src=".github/img/chat-interactive.gif" alt="devint chat interactive mode" width="1200"/>
  <p><i>Interactive mode with streaming responses and full conversation context</i></p>
</div>

Start an interactive chat session in a terminal UI that preserves conversation context throughout the session:

```bash
devint chat
```

## 📋 Command Reference <!-- omit in toc -->

| Flag                     | Type   | Required | Description                                                        |
| ------------------------ | ------ | -------- | ------------------------------------------------------------------ |
| [prompt]                 | string | ❌        | Prompt to send to the LLM (if omitted, starts interactive mode)    |
| --provider-override (-p) | string | ❌        | LLM provider override. Sets the LLM provider only for this request |
| --model-override (-m)    | string | ❌        | LLM model override. Sets the LLM model only for this request       |

# 🔧 Git Commands

The Developer Interface provides Git automation commands that leverage LLMs to streamline your Git workflow.

## devint git createpr <!-- omit in toc --> 

Automatically generates a comprehensive PR description from your git diff using an LLM and creates the PR on GitHub. This command eliminates the tedious task of manually writing PR descriptions by analyzing your code changes, commit history, and PR title to produce a well-structured description with primary and secondary changes.

**🛠️ Use Cases:**
- **Save Time**: Generate professional PR descriptions in seconds instead of spending minutes writing them manually
- **Consistency**: Ensure all PRs follow a consistent format and structure across your team
- **Context Awareness**: The LLM analyzes your diff, commits, and title to understand the "why" behind changes
- **Quality**: Automatically detects and highlights TODOs, primary functional changes, and secondary improvements

**💴 Value:** Reduces PR creation time from 5-10 minutes to under 30 seconds while improving description quality and consistency. Particularly valuable for teams that maintain high PR standards but want to reduce manual overhead.

| Flag                     | Type   | Required | Description                                                                         |
| ------------------------ | ------ | -------- | ----------------------------------------------------------------------------------- |
| --pr-title (-t)          | string | ✅        | PR title. Will open a draft PR if the string contains [DRAFT] or [WIP]              |
| --target-branch (-b)     | string | ❌        | Target branch (default "main")                                                      |
| --issue (-i)             | int    | ❌        | Issue number                                                                        |
| --dummy (-d)             | bool   | ❌        | Dummy mode. Will print summary to console and clipboard but not open a PR on GitHub |
| --provider-override (-p) | string | ❌        | LLM provider override. Sets the LLM provider only for this request                  |
| --model-override (-m)    | string | ❌        | LLM model override. Sets the LLM model only for this request                        |

## devint git summarizepr <!-- omit in toc -->

Generates an educational summary of a pull request by analyzing the PR diff with an LLM. The summary explains the rationale behind changes, highlights best practices, identifies potential issues, and provides learning insights. Perfect for code reviews, onboarding, and knowledge sharing.

**🛠️ Use Cases:**
- **Code Review Preparation**: Quickly understand complex PRs before reviewing, especially when working with unfamiliar codebases or languages
- **Learning & Onboarding**: Learn from existing PRs by understanding why changes were made and what patterns they demonstrate
- **Quality Assurance**: Get an LLM-powered second opinion on PRs, catching potential bugs, pitfalls, or optimization opportunities

**💴 Value:** Transforms PRs from code changes into educational resources. Instead of just seeing "what" changed, you understand "why" and "how" with context about best practices and potential issues. Saves time doing manual analysis and helps maintain institutional knowledge.

| Flag              | Type   | Required | Description                                                                    |
| ----------------- | ------ | -------- | ------------------------------------------------------------------------------ |
| --pr-number (-p)  | int    | ✅        | Pull request number to summarize                                               |
| --repo-owner (-r) | string | ❌        | GitHub repository owner override. If set, overrides the repo_owner from config |

**💡 Note:** The `pr_summary_output_dir` must be set in your `git_config` (via `devint config`) for this command to work.

# ⚙️ Configuration

Manage your Developer Interface configuration through the `devint config` command or by directly editing the configuration file.

## devint config <!-- omit in toc -->

| Flag          | Type   | Required | Description                                                                  |
| ------------- | ------ | -------- | ---------------------------------------------------------------------------- |
| --show (-s)   | bool   | ❌        | Show the configuration                                                       |
| --editor (-e) | string | ❌        | Edit the configuration in the given text editor, for example `nano` or `vim` |

To run the interactive editor, run without any flags, ie. `devint config`.

<!-- TODO: add GIF showing devint config interactive editor once UI is updated to use `github.com/rivo/tview` -->
<!-- <div align="center">
  <img src=".github/img/config-interactive.gif" alt="devint config interactive mode" width="800"/>
  <p><i>Interactive configuration editor</i></p>
</div> -->

## Configuration File <!-- omit in toc -->

The configuration is done through a config YAML file located at `~/.devint.config.yaml`.

An example configuration file can be found at `config/examples/.config.example.yaml`.

```yaml
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

# 📝 TODOs

*Development notes to improve the Developer Interface:*

- [ ] Update interactive Config interface to use `github.com/rivo/tview`
- [ ] Update interactive Config interface to log out OpenRouter models
- [ ] Add comprehensive table-driven tests for all packages

## 🧠 Additional GitHub Automation Ideas <!-- omit in toc -->

*🧠 Potential commands that could leverage LLMs to automate common GitHub workflows:*

- [ ] **`devint git reviewpr`**: Generate a code review checklist and suggestions for a PR, analyzing the diff for potential bugs, security issues, performance concerns, and code quality improvements
- [ ] **`devint git explaincommit`**: Explain what a specific commit does in plain language, useful for understanding complex changes or historical commits
- [ ] **`devint git generatechangelog`**: Analyze PRs merged since last release and generate a formatted changelog entry with categorized changes (features, fixes, breaking changes, etc.)
- [ ] **`devint git suggestlabels`**: Analyze PR title, description, and diff to suggest appropriate GitHub labels (e.g., `bug`, `feature`, `documentation`, `breaking-change`)
- [ ] **`devint git explainissue`**: Summarize and explain GitHub issues, breaking down complex problem descriptions into actionable insights
- [ ] **`devint git suggestreviewers`**: Analyze PR changes and suggest appropriate code reviewers based on file paths, languages, and change patterns
- [ ] **`devint git generatecommitmessage`**: Generate conventional commit messages from staged changes, useful for maintaining consistent commit history
- [ ] **`devint git analyzerelease`**: Compare two releases/tags and generate a comprehensive release notes document with categorized changes
- [ ] **`devint git explaindiff`**: Explain any arbitrary git diff in plain language, useful for understanding changes between branches or commits
- [ ] **`devint git suggestprtitle`**: Generate a descriptive PR title based on the diff and commit messages
