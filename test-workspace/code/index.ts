import postgres from "postgres";
import { Equal, FlexORM, GreaterThanOrEqual, type Users } from "./client";

const client = postgres("postgresql://postgres:postgres@localhost/postgres");
const orm = new FlexORM(client);

const obj = await orm.users.insert({
    name: "blah you go",
    created: null
})
.returning(orm.users.id, orm.users.name, orm.users.created);

console.log(obj);

const users = await orm.users
    .select()
    .where(new GreaterThanOrEqual(orm.users.id, 2))
    .limit(5) as Users[];

for (let user of users) {
    console.log(user.name);
}

client.end()