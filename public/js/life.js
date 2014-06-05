var width = 50,
	height = 50,
	clicking = false,
	clickMode = 0;

// The world data structure.

var world = {
	cells: [],
	width: width,
	height: height,
	init: function() {
		for (var y = 0; y < this.height; y++) {
			var row = [];

			for (var x = 0; x < this.width; x++) {
				h = row.push(false);
			}

			w = this.cells.push(row);
		}
		
		//console.log("cells: " + h + " rows (y) and " + w + " cells (x)");

		return this;
	},
	toggleCell: function(x, y) {
		this.cells[y][x] = !this.cells[y][x];
	},
	setCell: function(x, y, alive) {
		this.cells[y][x] = alive;
	},
	cell: function(x, y) {
		return this.cells[y][x];
	},
	updateWorldView: function($world) {
		el = $world[0];

		for (var y = 0; y < this.height; y++) {
			//if(!el.rows[y]) console.log("missing row " + y);

			row = el.rows[y];

			for (var x = 0; x < this.width; x++) {
				try {
					row.cells[x].className = this.cells[y][x] ? "alive" : "";
				}
				catch(e) {
					//console.log("Missing cell index " + x);
				}
				
			}
		}
	}
}.init();


// Create the view of the world.

$(document).ready(function() {
	var $world = $('#world');

	for (var y = 0; y < height; y++) {
		$row = $('<tr/>');

		for (var x = 0; x < width; x++) {
			$row.append('<td/>');
		}

		$world.append($row);
	};

	$('#world td').mousedown(function() {
		var $cell = $(this),
			x = $cell.index(),
			y = $cell.parent().index();

		$cell.toggleClass('alive');

		clickMode = !world.cell(x, y);
		world.setCell(x, y, clickMode);
		clicking = true;
	})
	.mouseenter(function() {
		if(clicking) {
			var $cell = $(this),
				x = $cell.index(),
				y = $cell.parent().index();

			if(clickMode) {
				$cell.addClass('alive');
			}
			else {
				$cell.removeClass('alive');
			}
			world.setCell(x, y, clickMode);
		}
		else return;
	})
	.attr('unselectable', 'on')
	.css('user-select', 'none')
	.on('selectstart', false);

	// Start button
	var ws = null;

	$('#start').click(function() {
		if(ws) {
			ws.close();
			ws = null;
			$(this).attr('value', 'Start');
			console.log('Stopping game');
		}
		else {
			$(this).attr('value', 'Stop');
			ws = newSim(width, height, $world);
			console.log(ws);
			console.log('Starting game');
		}
	});

	setInterval(function() {
		if (ws)
			console.log("recieved " + mcount, "server at " + scount);
	}, 1000);
})
.mouseup(function() {
	clicking = false;
});

// Request a new simulation

var mcount = 0;
var scount = 0;

function newSim(width, height, $world) {
	var ws = new WebSocket("ws://localhost:8080/game");

	ws.onopen = function() {
		ws.send(JSON.stringify({
			Command: "set",
			World: {
				Cells: world.cells,
				Width: width,
				Height: height
			}
		}));

		//console.log(world.cells);
	};

	ws.onmessage = function (e) {
		//console.log(e);

		try {
			data = $.parseJSON(e.data);
		}
		catch (e) {
			console.log("Unable to parse message");
			return
		}

		mcount += 1;

		//console.log(data);

		switch (data.Command) {
		case 'update':
	 		//console.log("Update recieved");
	 		world.cells = data.World
	 		world.updateWorldView($world);
	 		scount = data.SendCount;
		}
	};

	ws.onclose = function(e) {
		console.log("close:", e);
	};

	ws.onerror = function(e) {
		console.log("error:", e);
	};

	return ws;
}
