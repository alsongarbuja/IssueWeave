# IssueWeave 🕸️

> An AI-powered CLI tool that weaves together scattered GitHub issues into a single, cohesive timeline.

IssueWeave solves context-switching for maintainers of large open-source repositories. It scans issues, extracts cross-references, and uses AI to identify semantically related threads, weaving them into a single chronological view.

## ✨ Features

* **🔗 Explicit Weaving:** Automatically aggregates cross-referenced issues, pull requests, and inline comment mentions into a single timeline.
* **🧠 Semantic Weaving (WIP):** Leverages AI to analyze issue context and link related threads even when they aren't explicitly referenced.
* **📊 Web Visualization (Planned):** Generates a clean, interactive web page to visualize the woven issue tree.
* **🛠️ Data Manipulation (Planned):** Filter, sort, and export woven issue data.


## 🚀 Getting Started

> [!WARNING]
> IssueWeave is currently in early active development. Features and commands are subject to change.

### Prerequisites

* [Go](https://go.dev/doc/install) (1.20 or later)
* A GitHub Personal Access Token (Classic or Fine-Grained)

### Installation for Development

1. **Clone the repository:**
```bash
  git clone https://github.com/alsongarbuja/IssueWeave.git
  cd IssueWeave
```

2. **Set your GitHub Token:**
To avoid aggressive API rate limits, export your Personal Access Token.
```bash
export GITHUB_TOKEN="your_github_pat_here"
```

*(Note: AI provider API keys will be required for future semantic features).*

3. **Build the project:**
```bash
go build
```

## 💻 Usage

> [!NOTE]
> The examples below assume you have compiled the binary using `go build`. If you prefer running without building, replace `./issueweave` with `go run main.go`.

### `run`

The primary command to weave issues together. It fetches the root issue and chronologically threads all related references.

```bash
./issueweave run [owner/repo] [issueId]
```

**Arguments:**

| Position | Argument | Description |
| --- | --- | --- |
| 1 | `[owner/repo]` | **Required**. The repository target (e.g., `alsongabuja/IssueWeave`). |
| 2 | `[issueId]` | **Required**. The ID of the root issue to anchor the weave (e.g., `11234`). |

**Example:**

```bash
./issueweave run alsongarbuja/IssueWeave 333
```

### `rate`

A utility command to check your remaining GitHub API quota for both REST (Core) and GraphQL endpoints.

```bash
./issueweave rate
```

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!
If you're interested in helping build the AI semantic layer or the web visualization interface, feel free to open an issue to discuss the architecture.
