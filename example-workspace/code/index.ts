import postgres from "postgres";
import { FlexORM } from "./client";

const client = postgres("postgresql://postgres:postgres@localhost/postgres");
const orm = new FlexORM(client);

await orm.users.insert({
    name: "lincolnmaxwell"
});

const users = await orm.users.select();

for (let user of users) {
    console.log(user)
}

client.end()