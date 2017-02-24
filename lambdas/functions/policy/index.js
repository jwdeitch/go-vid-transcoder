exports.handle = function (e, ctx, cb) {

    // http://stackoverflow.com/a/1349426/4603498
    function makeid()
    {
        var text = "";
        var possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";

        for( var i=0; i < 10; i++ )
            text += possible.charAt(Math.floor(Math.random() * possible.length));

        return text;
    }

    // For security, we want to strip all non alpha-numeric characters
    function processFilename(rawFN) {
        if (rawFN.lastIndexOf(".") == -1) {
            return;
        }

        extension = rawFN.substring(rawFN.lastIndexOf("."), rawFN.length).replace(/\W/g, '');
        filename = rawFN.substring(0, rawFN.lastIndexOf(".")).replace(/[^a-z0-9+]+/gi, " ").trim().replace(/ /g, "_") + "#%#" + makeid();

        if (filename.length == 0) {
            filename = "I_did_not_provide_a_valid_filename"
        }

        if (extension.length == 0) {
            return false;
        }

        return filename + "." + extension;
    }

    var s3 = require("policyWriter");

    var s3Config = {
        accessKey: process.env.SAWS_ACCESS_KEY_ID,
        secretKey: process.env.SAWS_SECRET_ACCESS_KEY,
        bucket: process.env.WORKING_BUCKET,
        region: "us-east-1",
        type: e.type
    };

    cb(null, s3.s3Credentials(s3Config, "tmp/" + processFilename(e.file)));
};