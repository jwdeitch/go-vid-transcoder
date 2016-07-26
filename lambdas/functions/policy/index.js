exports.handle = function(e, ctx, cb) {

    var s3 = require( "policyWriter" );
    var env = require(".env.json");

    var s3Config = {
        accessKey: env.AWS_ACCESS_KEY_ID,
        secretKey: env.AWS_SECRET_ACCESS_KEY,
        bucket: env.WORKING_BUCKET,
        region: "us-east-1",
        type: e.type
    };

    cb(null, s3.s3Credentials(s3Config, "tmp/"+e.file));
};