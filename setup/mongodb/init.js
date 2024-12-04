db.createUser({
    user: "bogus",
    pwd: "bogus",
    roles: [{ role: "readWrite", db: "bogus" }],
});