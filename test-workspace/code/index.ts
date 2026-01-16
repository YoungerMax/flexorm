import postgres from "postgres";
import { Equal, FlexORM } from "./client";

const client = postgres("postgresql://postgres:postgres@localhost/postgres");
const orm = new FlexORM(client);

const result = await orm.users.insert({
    name: "lincolnmaxwell",
}).returning(orm.users.id);

const userId = result[0]['id'] as number;

await orm.posts.insert([
    { text: "Hello world!", author_id: userId },
    { text: "Isn't today a great day?!", author_id: userId }
]);

const postsResult = await orm.posts.select().leftJoin(orm.users, new Equal(orm.users.id, orm.posts.author_id));

console.log(postsResult);

client.end()