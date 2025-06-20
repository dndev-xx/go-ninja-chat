class Chat {
    constructor(id, clientId) {
        this.id = id;
        this.clientId = clientId;
    }

    static FromData(data) {
        return new Chat(data.chatId, data.clientId);
    }

    render() {
        return `
<div class="row card problem" data-chat-id="${this.id}">
    Chat with ${this.clientId}
</div>
`;
    }
}

class Message {
    constructor(id, authorId, body, createdAtStr) {
        this.id = id;
        this.authorId = authorId;
        this.body = body;
        this.createdAt = new Date(createdAtStr);

        if (!this.id) {
            console.warn("message id is undefined")
        }
        if (!this.authorId) {
            console.warn("message authorId is undefined")
        }
        if (!this.body) {
            console.warn("message body is empty")
        }
        if (!this.createdAt) {
            console.warn("message createdAt is undefined")
        }
    }

    static FromData(data) {
        return new Message(
            data.id || data.messageId,
            data.authorId,
            data.body,
            data.createdAt,
        );
    }

    render() {
        if (this.authorId === App.managerID) {
            return `
<div class="media media-chat media-chat-reverse" data-message-id="${this.id}">
    <div class="media-body">
        <div class="body-with-checks">
            <p>${this.body}</p>
            <i class="fa-solid fa-check-double status"></i>
        </div>
        <p class="meta">${this.createdAt.toLocaleString()}</p>
    </div>
</div>`;
        }

        return `
 <div class="media media-chat" data-message-id="${this.id}">
    <div class="companion">
        <img class="companion-avatar" src="https://img.icons8.com/color/36/000000/administrator-male.png">
        <div class="companion-name">${this.authorId.split('-')[0]}</div>
    </div>
    <div class="media-body">
        <p>${this.body}</p><p class="meta">${this.createdAt.toLocaleString()}</p>
    </div>
</div>`;
    }
}
