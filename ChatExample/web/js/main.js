define(
	"main",
	[
		"MessageList"
	],
	function(MessageList) {
		var ws = new WebSocket("ws://" + window.location.hostname + ":8080/websocket");
		var list = new MessageList(ws);
		ko.applyBindings(list);
	}
);
