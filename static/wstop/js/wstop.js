$(document).ready(function(){
    var memoryfreeseries;

   var ws = new WebSocket('ws://127.0.0.1:8080/wsmemoryprocess/');
		    var self = this;

		    ws.onmessage = function (event) {
			//console.log(event.data);
			tbl = JSON.parse(event.data);
			console.log(tbl);
			$('div[id=tbl]').empty();
			var e = $(document.createElement('table')).appendTo('div[id=tbl]').attr("class", "table");
			e.append("<tr><th>Name</th><th>Pid</th><th>VmRam</th><th>VmPeak</th><th>State</th></tr>");
			$.each(tbl, function(i, item) {
			    item = JSON.parse(item);
			    e.append("<tr><td>" + item.name + "</td><td>" + item.pid + "</td>ss<td>" + item.data + "</td><td>" + item.peak + "</td><td><button class=\"btn btn-success\">" + item.state + "</button></td></tr>");
			});
		    }


    $('#container').highcharts({
	chart : {
	    type : 'line',
	    events : {
		load : function() {

		    var connection = new WebSocket('ws://127.0.0.1:8080/wsmemoryprocessgraph/');
		    var self = this;

		    connection.onmessage = function (event) {
		    	var data = JSON.parse(event.data);         
		    	var series = self.series[0];

		    	$.each(data, function (i, item) {
		    	    var item = $.parseJSON(item);
		    	    var series = self.get(item.name);
		    	    //console.log(series);
		    	    if(series) { // series already exists
				var point = [(new Date()).getTime(),JSON.parse(item.y)];
				//console.log(point);
		    		series.addPoint(point, false);
		    	    } else { //  new series
		    		self.addSeries({
		    		    data: [],
				    id:item.name,
		    		    name:item.name
		    		});
		    	    }
		    	});
			self.redraw();
		    };

		}
	    }
	},
	title : {
	    text : false
	},
	xAxis : {
	    type : 'datetime',
	    minRange : 60 * 1000
	},
	yAxis : {
	    title : {
		text : false
	    }
	},
	legend : {
	    enabled : true
	},
	plotOptions : {
	    series : {
		threshold : 0,
		marker : {
		    enabled : false
		}
	    }
	},
	series : []
    });

// pie memoire

    $('#container-mem').highcharts({
	chart: {
	    type: 'pie',
	    events : {
	    	load : function() {
	    	    memData = this.series[0];
	    	    var ws = new WebSocket("ws://localhost:8080/wsmemoryconsograph/");
	    	    ws.onmessage = function(e) {
	    		var d = $.parseJSON(e.data);
	    		memData.setData(d);
	    	    };	
	    	}
	    },
	},
	title: {
	    text: 'Memory usage'
	},
	plotOptions: {
	    pie: {
		shadow: false,
		center: ['50%', '50%']
	    }
	},
	tooltip: {
	    valueSuffix: 'Gb'
	},
	series: [{
	    name: 'Memory usage',
	    data: [],
	    size: '80%',
	    innerSize: '60%',
	    dataLabels: {
		formatter: function () {
		    return this.y > 1 ? this.y + 'Gb'  : null;
		}
	    }
	}]
    });

});
