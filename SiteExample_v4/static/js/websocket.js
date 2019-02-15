var socket;

// Функция, вызываемая после загрузки документа
$(document).ready(function () {
    // Создаем сокет
    socket = new WebSocket('ws://' + window.location.host + '/ws/join?uname=' + $('#uname').text());
    // Обработчик функции получения сообщений
    socket.onmessage = function (event) {
        // Получаем данные в виде Json
        var data = JSON.parse(event.data);

        // Создаем элемент для верстки
        var li = document.createElement('li');

        console.log(data);
        
        // Определяем тип сообщения
        switch (data.Type) {
        case 0: // JOIN
            // Текущий ли юзер или сторонний
            if (data.User == $('#uname').text()) {
                li.innerText = 'You joined the chat room.';
            } else {
                li.innerText = data.User + ' joined the chat room.';
            }
            break;
        case 1: // LEAVE
            // Отключение от сервера
            li.innerText = data.User + ' left the chat room.';
            break;
        case 2: // MESSAGE
            // Создаем элемент
            var username = document.createElement('strong');
            var content = document.createElement('span');

            // Выставляем данные
            username.innerText = data.User;
            content.innerText = data.Content;

            // Добавляем к верстке
            li.appendChild(username);
            li.appendChild(document.createTextNode(': '));
            li.appendChild(content);

            break;
        }

        // Добавляем элемент перед остальными
        $('#chatbox li').first().before(li);
    };

    // Функция отправки сообщения
    var postConecnt = function () {
        // Получаем данные из верстки
        var uname = $('#uname').text();
        var content = $('#sendbox').val();
        // Отправляем сообщение
        socket.send(content);
        // Обнуляем данные в поле текста
        $('#sendbox').val('');
    }

    // Обработки клика на кнопку
    $('#sendbtn').click(function () {
        postConecnt();
    });
});
