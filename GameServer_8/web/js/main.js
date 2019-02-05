define(
	"main",
	[
	],
	function() {
        var app = new PIXI.Application(800, 600, {backgroundColor : 0x1099bb});
        document.body.appendChild(app.view);

        // create a texture from an image path
        var playerTexture = PIXI.Texture.fromImage('resources/ava.png');
        var ballTexture = PIXI.Texture.fromImage('resources/ava.png');
        playerTexture.baseTexture.scaleMode = PIXI.SCALE_MODES.NEAREST;
        ballTexture.baseTexture.scaleMode = PIXI.SCALE_MODES.NEAREST;

        var currentPlayerId = 0;
        var leftPlayer = null;
        var rightPlayer = null;
        var ball = null;

        var ws = new WebSocket("ws://" + window.location.hostname + ":8080/websocket");
        ws.onmessage = function(e) {
            var inputJson = $.evalJSON(e.data);

            // Инициализация юзера после подключения
            if(currentPlayerId == 0 && inputJson.messageType == 0){
                userState = null
                if(inputJson.leftPlayer.id > 0){
                    userState = inputJson.leftPlayer;
                }else if(inputJson.rightPlayer.id > 0){
                    userState = inputJson.rightPlayer;
                }

                if(userState != null){
                    newPlayer = createPlayer(userState.id, userState.t, userState.y, userState.h, userState.st, true);
                    currentPlayerId = userState.id;
                    
                    if (userState.t == 0){
                        this.leftPlayer = newPlayer;
                    }else{
                        this.rightPlayer = newPlayer;
                    }
    
                    // add it to the stage
                    app.stage.addChild(newPlayer);
                }
            }

            // Обновление состояния карты
            if(inputJson.messageType == 1){
                // Обновляем левого игрока
                if (inputJson.leftPlayer.id > 0 && inputJson.leftPlayer.id != currentPlayerId){
                    userState = inputJson.leftPlayer;
                    if (this.leftPlayer != null){
                        this.leftPlayer.y = userState.y;
                    }else{
                        this.leftPlayer = createPlayer(userState.id, userState.t, userState.y, userState.h, userState.st, false);
                        app.stage.addChild(this.leftPlayer);
                    }
                }

                // Обновляем правого игрока
                if (inputJson.rightPlayer.id > 0 && inputJson.rightPlayer.id != currentPlayerId){
                    userState = inputJson.rightPlayer;
                    if (this.rightPlayer != null){
                        this.rightPlayer.y = userState.y;
                    }else{
                        this.rightPlayer = createPlayer(userState.id, userState.t, userState.y, userState.h, userState.st, false);
                        app.stage.addChild(this.rightPlayer);
                    }
                }

                // Обновляем шар на поле
                roomState = inputJson.room
                if(this.ball != null){
                    this.ball.x = roomState.ballPosX;
                    this.ball.y = roomState.ballPosY;
                }else{
                    this.ball = createBall(roomState.ballPosX, roomState.ballPosY);
                    app.stage.addChild(this.ball);
                }
            }
        };

        function createBall(x, y) {
            // create our little bunny friend..
            var sprite = new PIXI.Sprite(ballTexture);

            sprite.x = x;
            sprite.y = y;

            // enable the bunny to be interactive... this will allow it to respond to mouse and touch events
            sprite.interactive = false;
            // this button mode will mean the hand cursor appears when you roll over the bunny with your mouse
            sprite.buttonMode = false;
            // center the bunny's anchor point
            sprite.anchor.set(0.5);
            // make it a bit bigger, so it's easier to grab
            sprite.scale.set(1);

            return sprite
        }

        function createPlayer(id, type, y, height, status, interactive) {
            // create our little bunny friend..
            var playerSprite = new PIXI.Sprite(playerTexture);

            playerSprite.id = id
            playerSprite.y = y;
            playerSprite.type = type
            playerSprite.height = height
            playerSprite.status = status

            if (type == 0){
                // move the sprite to its designated position
                playerSprite.x = 0;
            }else{
                // move the sprite to its designated position
                playerSprite.x = 800;
            }

            // enable the bunny to be interactive... this will allow it to respond to mouse and touch events
            playerSprite.interactive = interactive;
            // this button mode will mean the hand cursor appears when you roll over the bunny with your mouse
            playerSprite.buttonMode = true;
            // center the bunny's anchor point
            playerSprite.anchor.set(0.5);
            // make it a bit bigger, so it's easier to grab
            playerSprite.scale.set(1);

            if (interactive == false){
                playerSprite.alpha = 0.8;
            }

            // setup events for mouse + touch using
            // the pointer events
            playerSprite
                .on('pointerdown', onDragStart)
                .on('pointerup', onDragEnd)
                .on('pointerupoutside', onDragEnd)
                .on('pointermove', onDragMove);

            // For mouse-only events
            // .on('mousedown', onDragStart)
            // .on('mouseup', onDragEnd)
            // .on('mouseupoutside', onDragEnd)
            // .on('mousemove', onDragMove);

            // For touch-only events
            // .on('touchstart', onDragStart)
            // .on('touchend', onDragEnd)
            // .on('touchendoutside', onDragEnd)
            // .on('touchmove', onDragMove);

            return playerSprite
        }

        function onDragStart(event) {
            // store a reference to the data
            // the reason for this is because of multitouch
            // we want to track the movement of this particular touch
            this.data = event.data;
            this.alpha = 0.5;
            this.dragging = true;
        }

        function onDragEnd() {
            this.alpha = 1;
            this.dragging = false;
            // set the interaction data to null
            this.data = null;
        }

        function onDragMove() {
            if (this.dragging) {
                var newPosition = this.data.getLocalPosition(this.parent);
                //this.x = newPosition.x;
                this.y = newPosition.y;

                var messageToServer = new Object()
                messageToServer.id = this.id
                messageToServer.y = this.y

                var json = $.toJSON(messageToServer)
                ws.send(json)
            }
        }
	}
);
