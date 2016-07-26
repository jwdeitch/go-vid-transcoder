exports.handle = function (e, ctx, cb) {

    function processFilename(rawFN) {

        extension = rawFN.substring(rawFN.lastIndexOf("."), rawFN.length);
        filename = rawFN.substring(0, rawFN.lastIndexOf("."));
        return filename.replace(/[^a-z0-9+]+/gi, " ").trim().replace(/ /g, "_") + "." + extension.replace(/\W/g, '');
    }

    var s3 = require("policyWriter");
    var env = require(".env.json");

    var s3Config = {
        accessKey: env.AWS_ACCESS_KEY_ID,
        secretKey: env.AWS_SECRET_ACCESS_KEY,
        bucket: env.WORKING_BUCKET,
        region: "us-east-1",
        type: e.type
    };

    cb(null, s3.s3Credentials(s3Config, "tmp/" + processFilename(e.file)));
};