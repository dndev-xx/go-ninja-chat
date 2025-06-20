function main() {
    App.Run();
}

class App {
    static managerID;
    static managerToken;

    static apiClient;
    static currentChatHistoryCursor;
    static currentChatID;

    static openChats = $('#open-chats');
    static problemSelector = '.problem';
    static readyToProblemsBtn = $('#ready-to-problems-btn');
    static chatArea = $('#chat-content');
    static msgInput = $('#msgInput');
    static sendButton = $('#sendBtn');
    static problemResolvedButton = $('#problem-resolved-btn');

    static programScroll = false;

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

                this.managerToken = keycloak.token;
                console.info(keycloak);
                console.info(this.managerToken);

                if (!keycloak.hasResourceRole(keycloakClientRole, keycloakClientId)) {
                    throw new Error('RBAC: No required role for this user');
                }

                const payload = parseJWT(this.managerToken);
                this.managerID = payload.sub;
                console.info('manager id: ' + this.managerID);

                this.apiClient = new APIClient(this.managerToken);

                App.GetOpenProblems();
                App.GetReadyToProblemsAv();
                App.InitListeners();
            })
            .catch((err) => {
                alert('Failed to initialize keycloak: ' + err);
            });
    }

    static GetOpenProblems() {
        this.apiClient.getChats()
            .then((result) => {
                for (const c of result.chats) {
                    this.DisplayNewChat(c)
                }

                if (result.chats.length > 0) {
                    $(`*[data-chat-id="${result.chats[0].chatId}"]`).trigger('click');
                }
            })
            .catch((err) => {
                alert('Get chats with open problems error: ' + err);
            });
    }

    static GetReadyToProblemsAv() {
        this.apiClient.getFreeHandsBtnAvailability()
            .then((result) => {
                if (result.available) {
                    this.readyToProblemsBtn.removeClass('disabled');
                }
            })
            .catch((err) => {
                alert('Get "ready to problems" button availability: ' + err);
            });
    }

    static InitListeners() {
        App.OpenChatOnClick();
        App.ReadyToProblemsOnBtnClick();
        App.GetChatHistoryOnScroll();
        App.SendMessageOnBtnClick();
        App.ResolveProblemOnBtnClick();
    }

    static ReadyToProblemsOnBtnClick() {
        const app = this;
        this.readyToProblemsBtn.click(function () {
            app.apiClient.freeHands()
                .then(() => {
                    app.readyToProblemsBtn.addClass('disabled waiting');
                    app.readyToProblemsBtn.text('Waiting for problems...');
                })
                .catch((err) => {
                    alert('Send free hands signal error: ' + err);
                });
        });
    }

    static OpenChatOnClick() {
        const app = this;
        $(document.body).on('click', App.problemSelector, function () {
            const chatId = $(this).data('chat-id');
            if (App.currentChatID === chatId) {
                return
            }
            console.log('selected chat: ' + chatId);

            $(`*[data-chat-id="${App.currentChatID}"]`).removeClass('selected');
            $(`*[data-chat-id="${chatId}"]`).addClass('selected');
            App.currentChatID = chatId;

            app.chatArea.empty();
            app.currentChatHistoryCursor = '';
            App.GetLastChatMessages();
            app.problemResolvedButton.removeClass('disabled');
        });
    }

    static GetLastChatMessages() {
        const app = this;
        app.programScroll = true;

        this.apiClient.getChatHistory(app.currentChatID, app.currentChatHistoryCursor)
            .then((result) => {
                app.currentChatHistoryCursor = result.next;

                for (const m of result.messages) {
                    app.chatArea.prepend(Message.FromData(m).render());
                }

                app.chatArea.animate({
                    scrollTop: app.chatArea[0].scrollHeight,
                }, 1000, () => {
                    app.programScroll = false;
                });
            })
            .catch((err) => {
                alert('Get last messages error: ' + err);
            });
    }

    static GetChatHistoryOnScroll() {
        const app = this;

        app.chatArea.scroll(() => {
            if (app.chatArea.length.is && app.chatArea.length.is(':empty')) {
                return;
            }
            if (app.programScroll || (app.chatArea.scrollTop() !== 0)) {
                return;
            }

            if (!app.currentChatHistoryCursor) {
                alert('No more messages');
                return;
            }

            app.apiClient.getChatHistory(app.currentChatID, app.currentChatHistoryCursor)
                .then((result) => {
                    app.currentChatHistoryCursor = result.next;

                    for (const m of result.messages) {
                        app.chatArea.prepend(Message.FromData(m).render());
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
            if (!App.currentChatID) {
                alert('No chat selected!');
                return;
            }

            const msgText = app.msgInput.val();
            if (msgText === '') {
                return;
            }

            app.apiClient.sendMessage(App.currentChatID, msgText)
                .then((msg) => {
                    msg.body = msgText;

                    app.msgInput.val('');
                    app.DisplayNewMessage(msg);
                })
                .catch((err) => {
                    alert('Send message error: ' + err);
                });
        });
    }

    static ResolveProblemOnBtnClick() {
        const app = this;

        app.problemResolvedButton.click(function () {
            if (!App.currentChatID) {
                alert('No chat selected!');
                return;
            }

            app.apiClient.closeChat(App.currentChatID)
                .then(() => {
                    app.chatArea.empty();
                })
                .catch((err) => {
                    alert('Resolve problem error: ' + err);
                });
        });
    }

    static DisplayNewChat(chat) {
        this.openChats.append(Chat.FromData(chat).render());
    }

    static DisplayNewMessage(msg) {
        this.chatArea.append(Message.FromData(msg).render());
        this.chatArea.animate({
            scrollTop: this.chatArea[0].scrollHeight,
        }, 1000);
    }
}
