echo "Killing any existing processes..."
pkill -9 game-server
lsof -ti:8080 | xargs kill -9
lsof -ti:5173 | xargs kill -9
DATE=$(date +%Y%m%d%H%M%S)
cp mud-platform-backend/server.log mud-platform-backend/server.log.backup.$DATE
truncate -s 0 mud-platform-backend/server.log
