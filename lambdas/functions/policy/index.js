exports.handle = function (e, ctx, cb) {

    // For security, we want to strip all non alpha-numeric characters
    function processFilename(rawFN) {
        if (rawFN.lastIndexOf(".") == -1) {
            return;
        }

        extension = rawFN.substring(rawFN.lastIndexOf("."), rawFN.length).replace(/\W/g, '');
        filename = rawFN.substring(0, rawFN.lastIndexOf(".")).replace(/[^a-z0-9+]+/gi, " ").trim().replace(/ /g, "_");

        if (filename.length == 0) {
            filename = "I_did_not_provide_a_valid_filename"
        }

        if (extension.length == 0) {
            return false;
        }

        return filename + "." + extension;
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