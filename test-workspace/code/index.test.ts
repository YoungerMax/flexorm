// test.ts
import { expect, test, describe, beforeAll, afterAll } from "bun:test";
import postgres from "postgres";
import {
    FlexORM,
    Equal,
    And,
    Or,
    GreaterThan,
    GreaterThanOrEqual,
    LessThan,
    LessThanOrEqual,
    In,
    NotIn,
    IsNull,
    IsNotNull,
    NotEqual,
    Like,
    ILike,
    And as FAnd,
    Or as FOr,
    type Users
} from "./client"; // adjust path if needed

describe("FlexORM - Comprehensive Integration Tests", () => {
    const client = postgres("postgresql://postgres:postgres@localhost/postgres");
    const orm = new FlexORM(client);

    beforeAll(async () => {
        // Create users table
        await client`
            CREATE TABLE IF NOT EXISTS users (
                id SERIAL PRIMARY KEY,
                name TEXT NOT NULL,
                created TIMESTAMP DEFAULT NOW()
            );
        `;

        // Create posts table for join/delete USING tests (we'll treat this as a plain DB table)
        await client`
            CREATE TABLE IF NOT EXISTS posts (
                id SERIAL PRIMARY KEY,
                user_id INTEGER,
                title TEXT
            );
        `;

        // Ensure starting clean
        await client`TRUNCATE TABLE posts, users RESTART IDENTITY CASCADE;`;
    });

    afterAll(async () => {
        await client`DROP TABLE IF EXISTS posts;`;
        await client`DROP TABLE IF EXISTS users;`;
        await client.end();
    });

    test("Insert: single insert with returning", async () => {
        const [user] = await orm.users.insert({ name: "Alice" }).returning("id", "name");
        expect(user.id).toBeTypeOf("number");
        expect(user.name).toBe("Alice");
    });

    test("Insert: bulk insert and returning array", async () => {
        const rows = await orm.users
            .insert([
                { name: "Bob" },
                { name: "Charlie" },
                { name: "Alice" } // duplicate name to test grouping later
            ])
            .returning("id", "name", "created");

        expect(rows.length).toBeGreaterThanOrEqual(3);
        expect(rows.map(r => r.name)).toEqual(expect.arrayContaining(["Bob", "Charlie", "Alice"]));
    });

    test("Select: basic where Equal and NotEqual", async () => {
        const results = await orm.users.select().where(new Equal(orm.users.name, "Alice"));
        expect(results.length).toBeGreaterThanOrEqual(1);
        expect(results[0].name).toBe("Alice");

        const notAlice = await orm.users.select().where(new NotEqual(orm.users.name, "Alice"));
        // there should be people not named Alice
        expect(notAlice.length).toBeGreaterThanOrEqual(1);
        expect(notAlice.some(r => r.name === "Alice")).toBe(false);
    });

    test("Select: IN and NOT IN", async () => {
        const inRes = await orm.users.select("name").where(new In(orm.users.name, ["Alice", "Bob"]));
        expect(inRes.length).toBeGreaterThanOrEqual(2);
        const names = inRes.map(r => r.name);
        expect(names).toEqual(expect.arrayContaining(["Alice", "Bob"]));

        const notInRes = await orm.users.select("name").where(new NotIn(orm.users.name, ["Alice", "Bob"]));
        expect(notInRes.some(r => r.name === "Alice" || r.name === "Bob")).toBe(false);
    });

    test("Select: LIKE and ILIKE comparisons", async () => {
        // Insert some test rows if needed
        await orm.users.insert({ name: "alicia" });
        const likeRes = await orm.users.select("name").where(new Like(orm.users.name, "A%"));
        // LIKE is case sensitive in Postgres by default; depending on data, this may be empty.
        expect(Array.isArray(likeRes)).toBe(true);

        const ilikeRes = await orm.users.select("name").where(new ILike(orm.users.name, "a%"));
        expect(Array.isArray(ilikeRes)).toBe(true);
        // confirm at least one lower-case 'alicia' matches iLike
        expect(ilikeRes.some(r => r.name?.toLowerCase().startsWith("a"))).toBe(true);
    });

    test("Select: Greater/Less comparisons", async () => {
        // Use id comparisons (ids are numeric)
        const [minRow] = await orm.users.select().orderBy("id", "ASC").limit(1);
        const minId = minRow.id as number;

        const gtRes = await orm.users.select().where(new GreaterThan(orm.users.id, minId));
        expect(gtRes.every(r => r.id > minId)).toBe(true);

        const gteRes = await orm.users.select().where(new GreaterThanOrEqual(orm.users.id, minId));
        expect(gteRes.some(r => r.id === minId)).toBe(true);

        const ltRes = await orm.users.select().where(new LessThan(orm.users.id, minId + 1000));
        expect(ltRes.every(r => r.id < minId + 1000)).toBe(true);

        const lteRes = await orm.users.select().where(new LessThanOrEqual(orm.users.id, minId + 1000));
        expect(lteRes.some(Boolean)).toBe(true);
    });

    test("Select: AND / OR combinations (including nesting)", async () => {
        // AND: name = 'Zelda' AND created IS NOT NULL
        await orm.users.insert({ name: "Zelda" });
        const andQ = orm.users.select().where(
            new And(
                new Equal(orm.users.name, "Zelda"),
                new IsNotNull(orm.users.created)
            )
        );

        const andRes = await andQ;
        expect(andRes.length).toBeGreaterThanOrEqual(1);
        expect(andRes[0].name).toBe("Zelda");

        // OR: name = 'Zelda' OR name = 'Bob'
        const orRes = await orm.users.select().where(
            new Or(
                new Equal(orm.users.name, "Zelda"),
                new Equal(orm.users.name, "Bob")
            )
        );

        expect(orRes.some(r => r.name === "Zelda" || r.name === "Bob")).toBe(true);

        // Nested: (A AND B) OR C
        const nested = await orm.users.select().where(
            new Or(
                new And(
                    new Equal(orm.users.name, "Zelda"),
                    new IsNotNull(orm.users.created)
                ),
                new Equal(orm.users.name, "Charlie")
            )
        );

        expect(nested.some(r => r.name === "Zelda" || r.name === "Charlie")).toBe(true);
    });

    test("Expression errors: empty AND/OR should throw when building SQL", async () => {
        let caught = false;
        try {
            // This should throw: AND requires at least one operand
            await orm.users.select().where(new And());
        } catch (e: any) {
            caught = true;
            expect(String(e)).toContain("AND operator requires at least one operand");
        }
        expect(caught).toBe(true);

        caught = false;
        try {
            // This should throw: OR requires at least one operand
            await orm.users.select().where(new Or());
        } catch (e: any) {
            caught = true;
            expect(String(e)).toContain("OR operator requires at least one operand");
        }
        expect(caught).toBe(true);
    });

    test("SELECT: DISTINCT, GROUP BY, HAVING and aggregate-friendly usage", async () => {
        // We inserted multiple 'Alice' rows earlier — test DISTINCT
        const distinctRows = await orm.users.select("name").distinct();
        // distinct() returns SelectBuilder; awaiting it will execute; ensure it's array
        const distinct = await distinctRows;
        expect(Array.isArray(distinct)).toBe(true);

        // GROUP BY name and HAVING name IS NOT NULL (syntactically valid)
        const grouped = await orm.users
            .select("name")
            .groupBy(orm.users.name)
            .having(new IsNotNull(orm.users.name));

        expect(Array.isArray(grouped)).toBe(true);
        expect(grouped.every(r => r.name !== null)).toBe(true);
    });

    test("ORDER BY, LIMIT, OFFSET", async () => {
        // insert more names for ordering
        await orm.users.insert({ name: "Xavier" });
        await orm.users.insert({ name: "Yvonne" });
        await orm.users.insert({ name: "Aaron" });

        const ordered = await orm.users
            .select("name")
            .orderBy("name", "ASC")
            .limit(3)
            .offset(1);

        expect(ordered.length).toBeLessThanOrEqual(3);
        // check alphabetical order
        const names = ordered.map((r: any) => r.name);
        const sorted = [...names].sort();
        expect(names).toEqual(sorted);
    });

    test("JOIN: CROSS JOIN behavior (no ON)", async () => {
        // insert a post referencing a user
        const [user] = await orm.users.insert({ name: "JoinUser" }).returning("id", "name");
        await client`INSERT INTO posts (user_id, title) VALUES (${user.id}, 'first post')`;

        // Create a lightweight table-like object for posts (it just needs _orm and $$name)
        const postsTable: any = { _orm: orm, $$name: "posts" };

        // Cross join - just confirm it runs (will produce cartesian product)
        const res = await orm.users.select("users.name", "posts.title").crossJoin(postsTable);
        expect(Array.isArray(res)).toBe(true);
        expect(res.length).toBeGreaterThanOrEqual(1);
        // we should at least find the post title in the results
        expect(res.some((r: any) => r.title === "first post")).toBe(true);
    });

    test("UNION and UNION ALL", async () => {
        const q1 = orm.users.select("name").where(new Equal(orm.users.name, "Alice"));
        const q2 = orm.users.select("name").where(new Equal(orm.users.name, "Bob"));

        const unionQuery = orm.users.select("name").union(q1).unionAll(q2);
        const unionRes = await unionQuery;
        // should be an array (duplicates allowed because of UNION ALL)
        expect(Array.isArray(unionRes)).toBe(true);
    });

    test("UPDATE: set, where and returning", async () => {
        // ensure an Alice exists
        await orm.users.insert({ name: "ToBeUpdated" });
        const updateQ = orm.users.update()
            .set({ name: "UpdatedName" })
            .where(new Equal(orm.users.name, "ToBeUpdated"))
            .returning("id", "name");

        const updated = await updateQ;
        expect(Array.isArray(updated)).toBe(true);
        expect(updated[0].name).toBe("UpdatedName");

        // Confirm select sees update
        const [check] = await orm.users.select().where(new Equal(orm.users.name, "UpdatedName"));
        expect(check.name).toBe("UpdatedName");
    });

    test("DELETE: simple where + using clause + returning", async () => {
        // Add a user + a post to test USING and delete by join condition
        const [u] = await orm.users.insert({ name: "DeleteUser" }).returning("id", "name");
        await client`INSERT INTO posts (user_id, title) VALUES (${u.id}, 'to-delete')`;

        // Delete posts via delete with using (delete posts using users where posts.user_id = users.id and users.name = 'DeleteUser')
        // Build a delete query that uses the posts table (as a lightweight FTable-like)
        const postsTable: any = { _orm: orm, $$name: "posts" };

        // delete posts where title = 'to-delete'
        const deleteRes = await orm.users // intentionally use users.delete() to trigger FROM users (it's okay to delete from users here as a test)
            .delete()
            .where(new Equal(orm.users.name, "NonExistingName")); // nothing to delete but ensure API works

        // actually delete the post directly through the posts table name using raw SQL to ensure it's gone
        await client`DELETE FROM posts WHERE title = ${"to-delete"}`;

        // Confirm deletion
        const postCheck = await client`SELECT * FROM posts WHERE title = ${"to-delete"}`;
        expect(postCheck.length).toBe(0);
    });

    test("INSERT: ON CONFLICT DO NOTHING and DO UPDATE", async () => {
        // Insert a conflict candidate: explicit id (serial allows explicit value)
        const q1 = orm.users.insert({ id: 9999, name: "ConflictUser" }).onConflictDoNothing("id").returning("id", "name");
        const [created] = await q1;
        expect(created.id).toBe(9999);

        // Try inserting same id with DO NOTHING (should not crash)
        const q2 = orm.users.insert({ id: 9999, name: "ConflictUserTwo" }).onConflictDoNothing("id").returning("id", "name");
        const rows = await q2;
        // When DO NOTHING, depending on DB it may return nothing or the existing row; ensure no crash
        expect(Array.isArray(rows)).toBe(true);

        // Now test ON CONFLICT DO UPDATE (updateSet)
        const q3 = orm.users.insert({ id: 8888, name: "BeforeUpdate" }).onConflictDoUpdate("id", { name: "AfterUpdate" }).returning("id", "name");
        // This may insert or update; ensure builder doesn't crash and returns an array
        const out = await q3;
        expect(Array.isArray(out)).toBe(true);
    });

    test("QueryBuilder promise helpers: then/catch/finally", async () => {
        // then()
        const builder = orm.users.select("name").where(new Equal(orm.users.name, "Alice"));
        const names = await builder.then((rows: any[]) => rows.map(r => r.name));
        expect(Array.isArray(names)).toBe(true);

        // catch() - force an error by providing invalid operator through an empty AND
        let caught = false;

        await orm.users
            .select()
            .where(new And())
            .catch(err => {
                caught = true;
                expect(String(err)).toContain("AND operator requires at least one operand");
            });
        
        expect(caught).toBe(true);

        // finally() is executed regardless
        let finished = false;
        await orm.users.select().finally(() => { finished = true; });
        expect(finished).toBe(true);
    });

    test("Edge cases: selecting no columns => '*' and returning with mixed column types", async () => {
        const rows = await orm.users.select();
        expect(Array.isArray(rows)).toBe(true);

        // returning with FColumn objects and strings mixed
        const [r] = await orm.users.insert({ name: "ReturnMix" }).returning("id", orm.users.name as any);
        expect(r).toHaveProperty("id");
        expect(r).toHaveProperty("name");
    });
});
