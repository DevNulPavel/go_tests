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

        var currentPlayerId = 0
        var leftPlayer = null
        var rightPlayer = null
        var ball = null

        var ws = new WebSocket("ws://" + window.location.hostname + ":8080/websocket");
        var list = new UserStateUpdater(ws);
        ko.applyBindings(list);

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
                bunny.x = 0;
            }else{
                // move the sprite to its designated position
                bunny.x = 800;
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
                this.x = newPosition.x;
                this.y = newPosition.y;

                var newStateModel = new UserState(this).toModel()
                var json = $.toJSON(newStateModel)
                ws.send(json)
            }
        }

        function UserStateUpdater(ws) {
            var that = this;
            this.states = ko.observableArray();

            this.currentState = ko.observable(new UserState());

            this.send = function() {
                var model = this.currentState().toModel();
                json = $.toJSON(model)
                ws.send(json);
            };

            ws.onmessage = function(e) {
                var inputJson = $.evalJSON(e.data);

                // Инициализация юзера после подключения
                if(currentPlayerId == 0 && inputJson.messageType == 0){
                    if(inputJson.leftPlayer.id > 0){

                    }
                    if(inputJson.rightPlayer.id > 0){

                    }

                    newPlayer = createPlayer(userState.id, userState.t, userState.y, userState.h, userState.st, true)
                    currentPlayerId = userState.id
                    
                    if (userState.t = 0){
                        leftPlayer = newPlayer
                    }else{
                        rightPlayer = newPlayer
                    }

                    // add it to the stage
                    app.stage.addChild(newBunny);
                }

                // Обновление состояния карты
                if(inputJson.messageType == 1){

                }

                /*var activeIds = []
                for (var i = 0; i < states.length; i++){
                    var userState = states[i]
                    activeIds.push(userState.id.toString())

                    if (currentPlayerId == userState.id){
                        continue
                    }

                    // Update
                    bunny = players[userState.id]
                    if (bunny != null){
                        bunny.id = userState.id;
                        bunny.x = userState.x;
                        bunny.y = userState.y;
                    }else{
                        newBunny = createBunny(userState.id, userState.x, userState.y, false)
                        players[userState.id] = newBunny

                        // add it to the stage
                        app.stage.addChild(newBunny);
                    }
                }

                // Remove
                for(key in players){
                    if (currentPlayerId.toString() === key){
                        continue
                    }

                    var contains = false
                    for (var i = 0; i < activeIds.length; i++){
                        if(activeIds[i] === key){
                            contains = true;
                            break;
                        }
                    }

                    if (contains == false){
                        // add it to the stage
                        app.stage.removeChild(players[key]);
                        delete players[key]
                    }
                }*/
            };
        }
	}
);
