const msgWasBlockedAlertId = 'msg-was-blocked-alert';
const msgWasBlockedAlert = `<div id=#{blockedMsgAlertId} class="alert alert-danger">Сообщение не было доставлено менеджеру по
причине наличия в нём чувствительной информации</div>`;

class Message {
    constructor(id, authorId, body, createdAtStr, isReceived = false, isBlocked = false, isService = false) {
        console.log('Message constructor called with:', { id, authorId, body, createdAtStr, isReceived, isBlocked, isService });

        this.id = id;
        this.authorId = authorId;
        this.body = body;
        this.createdAt = new Date(createdAtStr);
        this.isReceived = isReceived;
        this.isBlocked = isBlocked;
        this.isService = isService;

        if (!this.id) {
            console.warn('message id is undefined');
        }
        if (!this.authorId) {
            console.warn('message authorId is undefined');
        }
        if (!this.body) {
            console.warn('message body is empty');
        }
        if (!this.createdAt || isNaN(this.createdAt.getTime())) {
            console.warn('message createdAt is undefined or invalid');
        }
    }

    static FromData(data) {
        console.log('FromData data:', data);

        const messageId = data.id || data.messageID || data.messageId;
        const authorId = data.authorId || data.authorID;

        console.log('Resolved messageId:', messageId);
        console.log('Resolved authorId:', authorId);

        return new Message(
            messageId,
            authorId,
            data.body,
            data.createdAt,
            data.isReceived || false,
            data.isBlocked || false,
            data.isService || false,
        );
    }

    render() {
        if (this.authorId === App.clientID) {
            let body = `<p class="body">${this.body}</p>`;
            if (this.isBlocked) {
                body = msgWasBlockedAlert;
            }

            const check = this.isReceived ? 'fa-check-double' : 'fa-check';

            return `
<div class="media media-chat media-chat-reverse" data-message-id="${this.id}">
    <div class="media-body">
        <div class="body-with-checks">
            ${body}
            <i class="fa-solid ${check} status"></i>
        </div>
        <p class="meta">${this.createdAt.toLocaleString()}</p>
    </div>
</div>`;
        }

        if (this.isService) {
            return `
 <div class="media media-chat" data-message-id="${this.id}">
    <div class="media-body">
        <div class="alert alert-secondary"><i class="fa-solid fa-circle-info"></i>&nbsp;${this.body}</div>
        <p class="meta">${this.createdAt.toLocaleString()}</p>
    </div>
 </div>
        `;
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