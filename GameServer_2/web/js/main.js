define(
	"main",
	[
	],
	function() {
        var app = new PIXI.Application(800, 600, {backgroundColor : 0x1099bb});
        document.body.appendChild(app.view);

        // create a texture from an image path
        var texture = PIXI.Texture.fromImage('resources/ava.png');

        // Scale mode for pixelation
        texture.baseTexture.scaleMode = PIXI.SCALE_MODES.NEAREST;

        var currentPlayerId = 0
        var bunnies = new Object()

        /*testBunny = createBunny(99, 200, 50)
        bunnies[99] = testBunny
        app.stage.addChild(testBunny);*/

        var ws = new WebSocket("ws://" + window.location.hostname + ":8080/websocket");
        var list = new UserStateUpdater(ws);
        ko.applyBindings(list);

        function createBunny(id, x, y, interactive) {

            // create our little bunny friend..
            var bunny = new PIXI.Sprite(texture);

            bunny.id = id

            // enable the bunny to be interactive... this will allow it to respond to mouse and touch events
            bunny.interactive = interactive;

            // this button mode will mean the hand cursor appears when you roll over the bunny with your mouse
            bunny.buttonMode = true;

            // center the bunny's anchor point
            bunny.anchor.set(0.5);

            // make it a bit bigger, so it's easier to grab
            bunny.scale.set(1);

            if (interactive == false){
                bunny.alpha = 0.8;
            }

            // setup events for mouse + touch using
            // the pointer events
            bunny
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

            // move the sprite to its designated position
            bunny.x = x;
            bunny.y = y;

            return bunny
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
                var states = []

                var model = $.evalJSON(e.data);
                for (var i = 0; i < model.length; i++) {
                    //console.log(model[i]);
                    var msg = new UserState(model[i]);
                    states.push(msg)
                }


                // Current user check
                if(currentPlayerId == 0){
                    var userState = states[states.length-1]

                    newBunny = createBunny(userState.id, userState.x, userState.y, true)
                    currentPlayerId = userState.id
                    bunnies[userState.id] = newBunny

                    // add it to the stage
                    app.stage.addChild(newBunny);
                }

                var activeIds = []
                for (var i = 0; i < states.length; i++){
                    var userState = states[i]
                    activeIds.push(userState.id.toString())

                    if (currentPlayerId == userState.id){
                        continue
                    }

                    // Update
                    bunny = bunnies[userState.id]
                    if (bunny != null){
                        bunny.id = userState.id;
                        bunny.x = userState.x;
                        bunny.y = userState.y;
                    }else{
                        newBunny = createBunny(userState.id, userState.x, userState.y, false)
                        bunnies[userState.id] = newBunny

                        // add it to the stage
                        app.stage.addChild(newBunny);
                    }
                }

                // Remove
                for(key in bunnies){
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
                        app.stage.removeChild(bunnies[key]);
                        delete bunnies[key]
                    }
                }
            };
        }

        function UserState(model) {
            if (model !== undefined) {
                this.id = model.id;
                this.x = model.x;
                this.y = model.y;
            } else {
                this.id = 0;
                this.x = 0;
                this.y = 0;
            }

            this.toModel = function() {
                return {
                    id: this.id,
                    x: this.x,
                    y: this.y
                };
            }
        }
	}
);
