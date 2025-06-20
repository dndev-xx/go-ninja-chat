const getChatsPath = '/getChats';
const getChatHistoryPath = '/getChatHistory';
const getFreeHandsBtnAvPath = '/getFreeHandsBtnAvailability';
const freeHandsPath = '/freeHands';
const sendMessagePath = '/sendMessage';
const closeChatPath = '/closeChat';

const defaultHistoryPageSize = 10;

class APIClient {
    constructor(token) {
        this.token = token;
    }

    async getFreeHandsBtnAvailability() {
        const response = await fetch(apiEndpoint + getFreeHandsBtnAvPath, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json;charset=utf-8',
                'Authorization': 'Bearer ' + this.token,
                'X-Request-ID': uuidV4(),
            },
        });
        return await this.extractData(response);
    }

    async freeHands() {
        const response = await fetch(apiEndpoint + freeHandsPath, {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + this.token,
                'X-Request-ID': uuidV4(),
            },
        });
        return await this.extractData(response);
    }

    async getChats() {
        const response = await fetch(apiEndpoint + getChatsPath, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json;charset=utf-8',
                'Authorization': 'Bearer ' + this.token,
                'X-Request-ID': uuidV4(),
            },
        });
        return await this.extractData(response);
    }

    async getChatHistory(chatId, cursor) {
        const request = {chatId};
        if (cursor) {
            request.cursor = cursor;
        } else {
            request.pageSize = defaultHistoryPageSize;
        }

        const response = await fetch(apiEndpoint + getChatHistoryPath, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json;charset=utf-8',
                'Authorization': 'Bearer ' + this.token,
                'X-Request-ID': uuidV4(),
            },
            body: JSON.stringify(request),
        });
        return await this.extractData(response);
    }

    async sendMessage(chatId, msgBody) {
        const response = await fetch(apiEndpoint + sendMessagePath, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json;charset=utf-8',
                'Authorization': 'Bearer ' + this.token,
                'X-Request-ID': uuidV4(),
            },
            body: JSON.stringify({
                chatId: chatId,
                messageBody: msgBody,
            }),
        });
        return await this.extractData(response);
    }

    async closeChat(chatId) {
        const response = await fetch(apiEndpoint + closeChatPath, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json;charset=utf-8',
                'Authorization': 'Bearer ' + this.token,
                'X-Request-ID': uuidV4(),
            },
            body: JSON.stringify({
                chatId: chatId,
            }),
        });
        return await this.extractData(response);
    }

    async extractData(response) {
        if (!response.ok) {
            throw new Error(`${response.status}`);
        }

        const result = await response.json();
        if (result.error) {
            throw new Error(`${result.error.code}: ${result.error.message}`);
        }
        return result.data;
    }
}
