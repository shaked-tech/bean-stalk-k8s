# Browser-Use Agent Quickstart Guide

This guide will help you set up and run your first browser automation agent using the browser-use library.

## Prerequisites

- Python 3.12 or higher
- An API key from one of the supported LLM providers (Gemini, OpenAI, or Anthropic)

## Step 1: Environment Setup

Choose one of the following methods to set up your Python environment:

### Option A: Using uv (Recommended)

```bash
# Create virtual environment with Python 3.12
uv venv --python 3.12

# Activate the environment (Unix/Linux/macOS)
source .venv/bin/activate

# Activate the environment (Windows)
.venv\Scripts\activate

# Install dependencies
uv pip install browser-use
uvx playwright install chromium --with-deps
```

### Option B: Using standard Python venv

```bash
# Create virtual environment
python3.12 -m venv .venv

# Activate the environment (Unix/Linux/macOS)
source .venv/bin/activate

# Activate the environment (Windows)
.venv\Scripts\activate

# Install dependencies
pip install browser-use
pip install playwright && playwright install chromium --with-deps
```

## Step 2: Configure Your LLM API Key

### Create Environment File

**Unix/Linux/macOS:**
```bash
touch .env
```

**Windows:**
```cmd
echo. > .env
```

### Add Your API Key

Choose one of the following providers and add the corresponding API key to your `.env` file:

**For Google Gemini (Free option available):**
```env
GEMINI_API_KEY=your_api_key_here
```
> Get a free Gemini API key: [Google AI Studio](https://aistudio.google.com/app/u/1/apikey?pli=1)

**For OpenAI:**
```env
OPENAI_API_KEY=your_api_key_here
```

**For Anthropic Claude:**
```env
ANTHROPIC_API_KEY=your_api_key_here
```

> **Note:** You only need one API key. See [Supported Models](/customize/supported-models) for more options.

## Step 3: Run Your First Agent

Create a Python file (e.g., `test_agent.py`) with one of the following examples:

### Example 1: Using Google Gemini

```python
from browser_use import Agent, ChatGoogle
from dotenv import load_dotenv
import asyncio

load_dotenv()

async def main():
    llm = ChatGoogle(model="gemini-flash-latest")
    task = "Find the number 1 post on Show HN"
    agent = Agent(task=task, llm=llm)
    await agent.run()

if __name__ == "__main__":
    asyncio.run(main())
```

### Example 2: Using OpenAI GPT

```python
from browser_use import Agent, ChatOpenAI
from dotenv import load_dotenv
import asyncio

load_dotenv()

async def main():
    llm = ChatOpenAI(model="gpt-4.1-mini")
    task = "Find the number 1 post on Show HN"
    agent = Agent(task=task, llm=llm)
    await agent.run()

if __name__ == "__main__":
    asyncio.run(main())
```

### Example 3: Using Anthropic Claude

```python
from browser_use import Agent, ChatAnthropic
from dotenv import load_dotenv
import asyncio

load_dotenv()

async def main():
    llm = ChatAnthropic(model='claude-sonnet-4-0', temperature=0.0)
    task = "Find the number 1 post on Show HN"
    agent = Agent(task=task, llm=llm)
    await agent.run()

if __name__ == "__main__":
    asyncio.run(main())
```

## Step 4: Run Your Agent

Execute your Python script:

```bash
python test_agent.py
```

The agent will automatically:
1. Open a browser
2. Navigate to Hacker News
3. Find the number 1 post on Show HN
4. Report back with the results

## Troubleshooting

- **Import Error:** Make sure you've activated your virtual environment
- **API Key Error:** Check that your `.env` file is in the same directory as your script
- **Browser Issues:** Ensure Chromium was installed correctly with `playwright install chromium --with-deps`
- **Permission Issues:** On Linux, you might need to install additional dependencies for Chromium

## Next Steps

- Explore different tasks and prompts
- Check out the [Supported Models](/customize/supported-models) documentation
- Learn about advanced agent configuration options
