const sendMessagePath = '/sendMessage';
const getHistoryPath = '/getHistory';

const defaultHistoryPageSize = 10;

class APIClient {
    constructor(token) {
        this.token = token;
    }

    async getHistory(cursor) {
        const request = {};
        if (cursor) {
            request.cursor = cursor;
        } else {
            request.pageSize = defaultHistoryPageSize;
        }

        const response = await fetch(apiEndpoint + getHistoryPath, {
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

    async sendMessage(msgBody) {
        const response = await fetch(apiEndpoint + sendMessagePath, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json;charset=utf-8',
                'Authorization': 'Bearer ' + this.token,
                'X-Request-ID': uuidV4(),
            },
            body: JSON.stringify({
                messageBody: msgBody,
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
