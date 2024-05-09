package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/cors"
	"io"
	"log"
	"net/http"
)

const llm = "llama3"
const baseUrl = "http://localhost:11434/"
const generateApiEndpoint = "api/generate"

const generateApi = baseUrl + generateApiEndpoint

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", promptHandler)

	handler := cors.Default().Handler(mux)

	log.Fatal(http.ListenAndServe("localhost:8080", handler))
}

func promptHandler(w http.ResponseWriter, r *http.Request) {
	userPrompt, err := parseRequest(r)
	if err != nil {
		fmt.Printf("error while parsing user request : %v\n", err)
	}

	req, err := createPrompt(userPrompt)
	if err != nil {
		fmt.Printf("error while creating prompt request : %v\n", err)
	}

	resp, err := sendPromptAndReceiveResponse(req)
	if err != nil {
		fmt.Printf("error while sending prompt to llm : %v\n", err)
	}

	promptResponse, err := extractPromptResponse(resp)
	if err != nil {
		fmt.Printf("error while extracting prompt response  : %v\n", err)
	}

	if promptResponse != "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(promptResponse))
	}
}

func extractPromptResponse(resp []byte) (string, error) {
	repBody := &ResponseBody{}

	err := json.Unmarshal(resp, repBody)
	if err != nil {
		fmt.Printf("error while unmarshalling response body : %v\n", err)
	}

	return repBody.Response, nil
}

func sendPromptAndReceiveResponse(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("error sending request: %v\n", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error reading response body: %v\n", err)
	}

	return body, nil
}

func createPrompt(userPrompt string) (*http.Request, error) {
	ctx := context.Background()
	prompt := &PromptBody{
		Model:  llm,
		Prompt: userPrompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(prompt)
	if err != nil {
		fmt.Printf("error while marshaling data : %v \n", err)
	}

	promptAsRequestBody := bytes.NewReader(jsonData)

	req, err := http.NewRequestWithContext(ctx, "POST", generateApi, promptAsRequestBody)
	if err != nil {
		fmt.Printf("error while creating request for llm : %v\n", err)
	}

	return req, nil
}

func parseRequest(r *http.Request) (string, error) {
	userPrompt, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("error while reading request body : %v\n", err)
	}

	return string(userPrompt), nil
}

type PromptBody struct {
	Model  string `json:"model,omitempty"`
	Prompt string `json:"prompt" json:"prompt,omitempty"`
	Stream bool   `json:"stream" json:"stream,omitempty"`
}

type ResponseBody struct {
	Model    string `json:"model,omitempty"`
	Response string `json:"response,omitempty"`
}
