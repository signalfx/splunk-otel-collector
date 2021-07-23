'use strict';
var signalFx = require('signalfx');

var token = 'YOUR SIGNALFX TOKEN'; // Replace with you token

var client = new signalFx.Ingest(token, {
  ingestEndpoint: "http://localhost:9080",
  enableAmazonUniqueId: false, // Set this parameter to `true` to retrieve and add Amazon unique identifier as dimension
  dimensions: {dyno_id: 'test.cust_dim'} // This dimension will be added to every datapoint and event
});

// Sent datapoints routine
var counter = 0;
function loop() {
  setTimeout(function () {
    console.log(counter);
    var timestamp = (new Date()).getTime();
    var gauges = [{metric: 'test.cpu', value: counter % 10, timestamp: timestamp}];
    var counters = [{metric: 'cpu_cnt', value: counter % 2, timestamp: timestamp}];

    // Send datapoint
    client.send({gauges: gauges, counters: counters});

    counter += 1;
    loop();
  }, 1000);
}

loop();
