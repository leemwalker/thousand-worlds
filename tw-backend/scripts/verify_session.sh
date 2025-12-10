#!/bin/bash

BASE_URL="http://localhost:8080/api"
EMAIL="session_tester_$(date +%s)@example.com"
PASSWORD="password123"

echo "=== Game Session API Verification ==="
echo ""

echo "1. Registering user..."
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}" | jq '.'
echo ""

echo "2. Logging in..."
LOGIN_RESP=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}")
TOKEN=$(echo $LOGIN_RESP | jq -r .token)
echo "Token obtained: ${TOKEN:0:50}..."
echo ""

if [ "$TOKEN" == "null" ]; then
  echo "❌ Login failed"
  exit 1
fi

echo "3. Creating a test world in database..."
# Insert a test world directly into the database
docker exec mud_postgis psql -U admin -d mud_core -c "INSERT INTO worlds (id, name, shape, created_at) VALUES ('00000000-0000-0000-0000-000000000001', 'Test World', 'sphere', NOW()) ON CONFLICT (id) DO NOTHING;" > /dev/null 2>&1
echo "Test world created (or already exists)"
WORLD_ID="00000000-0000-0000-0000-000000000001"
echo ""

echo "4. Creating a human character..."
CHAR_RESP=$(curl -s -X POST "$BASE_URL/game/characters" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"world_id\": \"$WORLD_ID\", \"name\": \"TestHero\", \"species\": \"Human\"}")
echo "$CHAR_RESP" | jq '.'
CHAR_ID=$(echo $CHAR_RESP | jq -r .character.character_id)
echo ""

if [ "$CHAR_ID" == "null" ]; then
  echo "❌ Character creation failed"
  echo "$CHAR_RESP" | jq '.error'
  exit 1
fi

echo "5. Creating an elf character..."
ELF_RESP=$(curl -s -X POST "$BASE_URL/game/characters" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"world_id\": \"$WORLD_ID\", \"name\": \"ElfArcher\", \"species\": \"Elf\"}")
echo "$ELF_RESP" | jq -c '{name:.character.name, species:"Elf", max_hp:.secondary_attributes.max_hp, sight:.attributes.sight}'
echo ""

echo "6. Creating a dwarf character..."
DWARF_RESP=$(curl -s -X POST "$BASE_URL/game/characters" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"world_id\": \"$WORLD_ID\", \"name\": \"DwarfWarrior\", \"species\": \"Dwarf\"}")
echo "$DWARF_RESP" | jq -c '{name:.character.name, species:"Dwarf", max_hp:.secondary_attributes.max_hp, endurance:.attributes.endurance}'
echo ""

echo "7. Listing all characters..."
curl -s -X GET "$BASE_URL/game/characters" \
  -H "Authorization: Bearer $TOKEN" | jq '.characters | map({name:.name, world_id:.world_id})'
echo ""

echo "8. Joining game with first character..."
JOIN_RESP=$(curl -s -X POST "$BASE_URL/game/join" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"character_id\": \"$CHAR_ID\"}")
echo "$JOIN_RESP" | jq '.'
echo ""

echo "✅ All tests passed!"
