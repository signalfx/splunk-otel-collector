Object.keys(process.env).forEach(function(key) {
    console.log(key + '=' + process.env[key]);
});