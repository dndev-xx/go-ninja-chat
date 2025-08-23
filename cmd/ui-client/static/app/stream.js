const eventHandlers = {
    'NewMessageEvent': (event) => {
        if ($(`*[data-message-id="${event.messageId}"]`).length > 0) {
            return;
        }
        App.DisplayNewMessage(event);
    },

    'MessageSentEvent': (event) => {
        $(`*[data-message-id="${event.messageId}"]`).find('.status')
            .removeClass('fa-check')
            .addClass('fa-check-double');
    },

    'MessageBlockedEvent': (event) => {
        const msg = $(`*[data-message-id="${event.messageId}"]`);
        if (msg.find('#' + msgWasBlockedAlertId).length > 0) {
            return;
        }
        msg.find('.body').remove();
        msg.find('.body-with-checks').prepend(msgWasBlockedAlert);
    }
};

function initWsStream(token) {
    const sock = new WebSocket(wsEndpoint, [wsProtocol, token]);

    window.addEventListener('unload', function () {
        if (sock.readyState === WebSocket.OPEN) {
            sock.close();
        }
    });

    sock.onopen = function () {
        console.info('ws: connection established');
    };

    sock.onclose = function (event) {
        if (!event.wasClean) {
            console.error('ws: unexpected connection lost');
            console.error('code: ' + event.code + ', reason: ' + event.reason);
        }
    };

    sock.onerror = function (event) {
        console.error('ws: error: ' + JSON.stringify(event));

        // If error occurred then try to reconnect.
        (async () => {
            let promise = new Promise(resolve => setTimeout(resolve, 2000));

            await promise;

            initWsStream(token);
        })();
    };

    sock.onmessage = function (event) {
        console.info('ws: new event: ' + event.data);

        const payload = JSON.parse(event.data);
        const eventType = payload.eventType;

        if (!(eventType in eventHandlers)) {
            console.error('ws: unknown event: ' + eventType);
            return;
        }

        eventHandlers[eventType](payload);
    };
}
