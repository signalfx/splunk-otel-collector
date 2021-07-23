// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
