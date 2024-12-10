# QuizMeh 🎓🤖

Designed for fun, learning, and engagement, this bot allows you to test your knowledge with True/False and Multiple-Choice questions and tracks your score. 

I used aistudio to parse potential final exam questions to a json file! I would advice doing the same! 😊

# Features ✨
Interactive Quizzes:
* Supports True/False and Multiple-Choice questions.
* Easy-to-use commands for generating quizzes.
* Random or Sequential Questions:
* Choose to answer questions in order or let the bot pick randomly.


Immediate Feedback:
* Tells you if your answer is correct or wrong.
* Displays the correct answer for incorrect responses.

Commands 🛠️
```
!quiz [n]	Start a quiz with the first n questions.
!quiz random [n]	Start a quiz with n random questions.
```

# Installation 💻

## Prerequisite 👏
```
go-lang
```

1. Clone the Repository
```bash
git clone https://github.com/CharlesTheGreat77/QuizMeh
cd QuizMeh
```
2. Add the Environment Variable

Set your bot token as an environment variable:

On Linux/macOS:
```bash
export DISCORD_TOKEN="your-bot-token-here"
```
On Windows (Command Prompt):
```
set DISCORD_TOKEN=your-bot-token-here
```
3. Add Your Questions

## Quiz question/answer(s)
Place your final.json file with the following structure in the root directory:
```json
{
  "Exam": {
    "true_false": [
      {
        "question": "Sample True/False Question?",
        "choices": {
          "A": "True",
          "B": "False"
        },
        "answer": "A"
      }
    ],
    "multiple_choice": [
      {
        "question": "Sample Multiple Choice Question?",
        "choices": {
          "A": "Option A",
          "B": "Option B",
          "C": "Option C",
          "D": "Option D"
        },
        "answer": "B"
      }
    ]
  }
}
```

4. Run the Bot 🚀
```bash
go run geobot.go
```

# Contributing 🤝
Contributions are welcome! Feel free to open issues or submit pull requests.

# Acknowledgments 🙌
Special thanks to:
* DiscordGo for making Discord bot development easy.