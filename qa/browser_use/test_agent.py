from browser_use import Agent, ChatAnthropic
from dotenv import load_dotenv
import asyncio

load_dotenv()

async def main():
    llm = ChatAnthropic(model='claude-3-5-sonnet-20241022', temperature=0.0)
    task = "Find the number 1 post on Show HN"
    agent = Agent(task=task, llm=llm)
    await agent.run()

if __name__ == "__main__":
    asyncio.run(main())
