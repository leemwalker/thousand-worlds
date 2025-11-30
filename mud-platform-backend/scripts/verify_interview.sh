#!/bin/bash

BASE_URL="http://localhost:8080/api"
EMAIL="tester_$(date +%s)@example.com"
PASSWORD="password123"

echo "1. Registering user..."
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}"
echo -e "\n"

echo "2. Logging in..."
LOGIN_RESP=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}")
TOKEN=$(echo $LOGIN_RESP | jq -r .token)
echo "Token: $TOKEN"
echo -e "\n"

if [ "$TOKEN" == "null" ]; then
  echo "Login failed"
  exit 1
fi

echo "3. Starting interview..."
START_RESP=$(curl -s -X POST "$BASE_URL/world/interview/start" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{}")
echo "Response: $START_RESP"
SESSION_ID=$(echo $START_RESP | jq -r .session_id)
QUESTION=$(echo $START_RESP | jq -r .question)
echo "Session ID: $SESSION_ID"
echo "Question: $QUESTION"
echo -e "\n"

if [ "$SESSION_ID" == "null" ]; then
  echo "Start interview failed"
  exit 1
fi

echo "4. Sending message..."
MSG_RESP=$(curl -s -X POST "$BASE_URL/world/interview/message" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"session_id\": \"$SESSION_ID\", \"message\": \"I want a sci-fi world with high tech.\"} ")
NEXT_QUESTION=$(echo $MSG_RESP | jq -r .question)
echo "Next Question: $NEXT_QUESTION"
echo -e "\n"

echo "5. Getting active interview..."
ACTIVE_RESP=$(curl -s -X GET "$BASE_URL/world/interview/active" \
  -H "Authorization: Bearer $TOKEN")
ACTIVE_SESSION_ID=$(echo $ACTIVE_RESP | jq -r .session_id)
echo "Active Session ID: $ACTIVE_SESSION_ID"

if [ "$ACTIVE_SESSION_ID" == "$SESSION_ID" ]; then
  echo "Active session matches!"
else
  echo "Active session mismatch or not found"
fi
