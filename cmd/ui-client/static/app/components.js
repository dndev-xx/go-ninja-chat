const blockedMsgBody = `<div class="alert alert-danger">Сообщение не было доставлено менеджеру по
причине наличия в нём чувствительной информации</div>`;

class Message {
    constructor(id, authorId, body, createdAtStr, isReceived, isBlocked, isService) {
        this.id = id;
        this.authorId = authorId;
        this.body = body;
        this.createdAt = new Date(createdAtStr);
        this.isReceived = isReceived;
        this.isBlocked = isBlocked;
        this.isService = isService;
    }

    static FromData(data) {
        return new Message(
            data.id,
            data.authorId,
            data.body,
            data.createdAt,
            data.isReceived,
            data.isBlocked,
            data.isService,
        );
    }

    render() {
        if (this.authorId === App.clientID) {
            let body = `<p class="body">${this.body}</p>`;
            if (this.isBlocked) {
                body = blockedMsgBody;
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
        `
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
