package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Question struct {
	Question string            `json:"question"`
	Choices  map[string]string `json:"choices"`
	Answer   string            `json:"answer"`
}

func main() {
	botToken := os.Getenv("DISCORD_TOKEN")
	if botToken == "" {
		log.Fatal("Bot token is not set in the environment variable 'DISCORD_BOT_TOKEN'")
	}

	questions, err := loadQuestions("final.json")
	if err != nil {
		log.Fatalf("Error loading questions: %v", err)
	}

	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID || !strings.HasPrefix(m.Content, "!quiz") {
			return
		}
		handleQuizCommand(s, m, questions)
	})

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error connecting to Discord: %v", err)
	}
	fmt.Println("Bot is running. Press CTRL+C to exit.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	dg.Close()
}

func loadQuestions(filename string) ([]Question, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var wrapper struct {
		Exam struct {
			TrueFalse      []Question `json:"true_false"`
			MultipleChoice []Question `json:"multiple_choice"`
		} `json:"Exam"`
	}

	err = json.Unmarshal(data, &wrapper)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return append(wrapper.Exam.TrueFalse, wrapper.Exam.MultipleChoice...), nil
}

func handleQuizCommand(s *discordgo.Session, m *discordgo.MessageCreate, questions []Question) {
	args := strings.Fields(m.Content)
	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !quiz [n] or !quiz random [n]")
		return
	}

	numQuestions, err := strconv.Atoi(args[len(args)-1])
	if err != nil || numQuestions <= 0 {
		s.ChannelMessageSend(m.ChannelID, "Invalid number of questions.")
		return
	}

	var selectedQuestions []Question
	if len(args) > 2 && args[1] == "random" { // !quiz random n
		selectedQuestions = getRandomQuestions(questions, numQuestions)
	} else {
		if numQuestions > len(questions) {
			numQuestions = len(questions)
		}
		selectedQuestions = questions[:numQuestions]
	}

	s.ChannelMessageDelete(m.ChannelID, m.ID)

	go runQuiz(s, m.ChannelID, selectedQuestions)
}

func getRandomQuestions(questions []Question, n int) []Question {
	if n > len(questions) {
		n = len(questions)
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(questions), func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
	})
	return questions[:n]
}

func runQuiz(s *discordgo.Session, channelID string, questions []Question) {
	correct := 0
	var questionMessageIDs []string

	for _, q := range questions {
		correctAnswer := askQuestion(s, channelID, q, &questionMessageIDs)
		if correctAnswer {
			correct++
		}
	}

	finalMessage, _ := s.ChannelMessageSend(channelID, fmt.Sprintf("Quiz complete! You got %d/%d correct.", correct, len(questions)))
	time.AfterFunc(3*time.Second, func() {
		for _, msgID := range questionMessageIDs {
			s.ChannelMessageDelete(channelID, msgID) // delete quiz custer
		}
		s.ChannelMessageDelete(channelID, finalMessage.ID)
	})
}

func askQuestion(s *discordgo.Session, channelID string, q Question, messageIDs *[]string) bool {
	choices := make([]discordgo.MessageComponent, 0, len(q.Choices))
	for key, value := range q.Choices {
		if len(value) > 80 { // truncate if question(s) are big asf
			value = value[:80]
		}
		choices = append(choices, discordgo.Button{
			Label:    value,
			CustomID: key,
			Style:    discordgo.PrimaryButton,
		})
	}

	msg, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: q.Question,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{Components: choices},
		},
	})
	if err != nil {
		log.Printf("Failed to send question: %v", err)
		return false
	}

	*messageIDs = append(*messageIDs, msg.ID)

	answerChan := make(chan string)
	s.AddHandlerOnce(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Message == nil || i.Message.ID != msg.ID {
			return
		}

		selected := i.MessageComponentData().CustomID

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
		if err != nil {
			log.Printf("Failed to acknowledge interaction: %v", err)
			return
		}

		answerChan <- selected
	})

	select {
	case answer := <-answerChan:
		if answer == q.Answer {
			resp, _ := s.ChannelMessageSend(channelID, "Correct!")
			*messageIDs = append(*messageIDs, resp.ID)
			return true
		} else {
			resp, _ := s.ChannelMessageSend(channelID, fmt.Sprintf("Wrong! The correct answer was: %s", q.Choices[q.Answer]))
			*messageIDs = append(*messageIDs, resp.ID)
			return false
		}
	case <-time.After(20 * time.Second):
		resp, _ := s.ChannelMessageSend(channelID, "Time's up! Moving to the next question.")
		*messageIDs = append(*messageIDs, resp.ID)
		return false
	}
}
