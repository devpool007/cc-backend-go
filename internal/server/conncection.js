const ws = new WebSocket('ws://localhost:3000/ws');

ws.onopen = () => {
    console.log('Connected to WebSocket');
    // Join a room
    ws.send(JSON.stringify({
        type: 'join-room',
        roomId: 'room1',
        userId: 'user1'
    }));
};

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    console.log('Received message:', message);
};

ws.onclose = () => {
    console.log('Disconnected from WebSocket');
};

