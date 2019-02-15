var lastReceived = 0;
var isWait = false;

var fetch = function () {
    // Блокировка запроса
    if (isWait) {
        return;
    }
    isWait = true;

    // Выполняем запрос на сервер последних сообщений
    $.getJSON("/lp/fetch?lastReceived=" + lastReceived, function (data) {
        if (data == null) {
            return;
        }
        // Обходим все сообщения
        $.each(data, function (i, event) {
            // Создаем новый элемент
            var li = document.createElement('li');

            switch (event.Type) {
            case 0: // JOIN
                // Текущий ли юзер или сторонний
                if (event.User == $('#uname').text()) {
                    li.innerText = 'You joined the chat room.';
                } else {
                    li.innerText = event.User + ' joined the chat room.';
                }
                break;
            case 1: // LEAVE
                // Отключение от сервера
                li.innerText = event.User + ' left the chat room.';
                break;
            case 2: // MESSAGE
                // Создаем элемент
                var username = document.createElement('strong');
                var content = document.createElement('span');

                // Выставляем данные
                username.innerText = event.User;
                content.innerText = event.Content;

                // Добавляем к верстке
                li.appendChild(username);
                li.appendChild(document.createTextNode(': '));
                li.appendChild(content);

                break;
            }

            // Добавляем элемент перед остальными
            $('#chatbox li').first().before(li);

            lastReceived = event.Timestamp;
        });
        isWait = false;
    });
}

// Настраиваем периодичность получения сообщений
setInterval(fetch, 3000);

// Выполняем получение данных
fetch();

// Функция, устанавливая после загрузки
$(document).ready(function () {
    // Функция отправки сообщений
    var postMessage = function () {
        var uname = $('#uname').text();
        var content = $('#sendbox').val();
        $.post("/lp/post", {
            uname: uname,
            content: content
        });
        $('#sendbox').val("");
    }

    $('#sendbtn').click(function () {
        postMessage();
    });
});
