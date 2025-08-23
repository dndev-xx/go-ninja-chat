function main() {
    App.Run();
}

class App {
    static clientID;
    static clientToken;

    static apiClient;
    static historyCursor;

    static chatArea = $('#chat-content');
    static msgSelector = '.media.media-chat';
    static msgInput = $('#msgInput');
    static sendButton = $('#sendBtn');

    static Run() {
        const keycloak = new Keycloak({
            url: keycloakEndpoint,
            realm: keycloakRealm,
            clientId: keycloakClientId,
        });

        keycloak.init({
            onLoad: 'login-required',
        })
            .then((authenticated) => {
                if (!authenticated) {
                    throw new Error('Not authenticated!');
                }

                this.clientToken = keycloak.token;
                console.info(keycloak);
                console.info(this.clientToken);

                if (!keycloak.hasResourceRole(keycloakClientRole, keycloakClientId)) {
                    throw new Error('RBAC: No required role for this user');
                }

                const payload = parseJWT(this.clientToken);
                this.clientID = payload.sub;
                console.info('client id: ' + this.clientID);

                this.apiClient = new APIClient(this.clientToken);

                initWsStream(this.clientToken);
                App.GetLastMessages();
                App.InitListeners();
            })
            .catch((err) => {
                alert('Failed to initialize keycloak: ' + err);
            });
    }

    static GetLastMessages() {
        this.apiClient.getHistory(this.historyCursor)
            .then((result) => {
                this.historyCursor = result.next;

                for (const m of result.messages.reverse()) {
                    this.chatArea.append(Message.FromData(m).render());
                }

                this.chatArea.animate({
                    scrollTop: this.chatArea[0].scrollHeight,
                }, 1000);
            })
            .catch((err) => {
                alert('Get last messages error: ' + err);
            });
    }

    static InitListeners() {
        App.GetHistoryOnScroll();
        App.SendMessageOnBtnClick();
    }

    static GetHistoryOnScroll() {
        const app = this;
        this.chatArea.scroll(() => {
            if (app.chatArea.scrollTop() !== 0) {
                return;
            }

            if (!app.historyCursor) {
                alert('No more messages');
                return;
            }

            app.apiClient.getHistory(this.historyCursor)
                .then((result) => {
                    app.historyCursor = result.next;

                    for (const m of result.messages) {
                        $(Message.FromData(m).render()).insertBefore($(app.msgSelector).first());
                    }
                })
                .catch((err) => {
                    alert('Get the next messages error: ' + err);
                });
        });
    }

    static SendMessageOnBtnClick() {
        const app = this;
        this.sendButton.click(function () {
            const msgBody = app.msgInput.val();
            if (msgBody === '') {
                return;
            }

            app.apiClient.sendMessage(msgBody)
                .then((msg) => {
                    msg.body = msgBody;

                    app.msgInput.val('');
                    app.DisplayNewMessage(msg);
                })
                .catch((err) => {
                    alert('Send message error: ' + err);
                });
        });
    }

    static DisplayNewMessage(msg) {
        this.chatArea.append(Message.FromData(msg).render());
        this.chatArea.animate({
            scrollTop: this.chatArea[0].scrollHeight,
        }, 1000);
    }
}
