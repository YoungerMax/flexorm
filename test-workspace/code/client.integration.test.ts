import { $ } from "bun";
import { afterEach, beforeEach, describe, expect, it } from "bun:test";
import postgres from "postgres";
import { Equal, FlexORM, In, Line, Point, type ValidatorTestTable } from "./client";


describe("FlexORM Integration Tests", () => {
  let sql: postgres.Sql<{}>;
  let orm: FlexORM;
  const postgresUrl = 'postgres://postgres:postgres@localhost/postgres?sslmode=disable';
  const validData: ValidatorTestTable = {
      required_integer: 42,
      required_varchar_maxlength: "test value",
      required_char_fixed: "abcde",
      required_text_minlength: "minimum length text",
      required_uuid: "550e8400-e29b-41d4-a716-446655440000",
      required_json: { key: "value", number: 123 },
      required_boolean: true,
      required_numeric: 123,
      required_real: 3.14,
      required_double_precision: 2.718281828459045,
      required_timestamp: new Date(),
      required_date: new Date(),
      required_time: "14:30:00",
      required_enum: "OPTION_A",
      required_point: new Point(10.5, -5.2),
      required_line: new Line(1, -2, 3),
      required_interval: "2 days 4 hours",
    };

  beforeEach(async () => {
    // Setup test database connection
    sql = postgres(postgresUrl);
    orm = new FlexORM(sql);

    // Create table if it doesn't exist
    try {
      const result = await $`cd ../database && ./cli migrations up --to 0 --database-url="${postgresUrl}"`.text();
      console.log(result);
    } catch (error) {
      console.error("Failed to create table:", error);
    }

    // Clean up before each test
    try {
      await sql`TRUNCATE TABLE validator_test_table RESTART IDENTITY CASCADE`;
    } catch (error) {
      // Ignore cleanup error
    }
  });

  afterEach(async () => {
    await sql.end();
  });

  describe("Database Operations", () => {
    it("should insert and retrieve data", async () => {
      // Insert data
      const insertResult = await orm.validator_test_table
        .insert(validData)
        .returning(orm.validator_test_table.id) as Array<Partial<ValidatorTestTable>>;

      expect(insertResult).toBeDefined();
      expect(Array.isArray(insertResult)).toBe(true);

      // Select and verify data
      const selectedData = await orm.validator_test_table
        .select()
        .where(new Equal(orm.validator_test_table.id, insertResult[0]?.id)) as Array<ValidatorTestTable>;

      expect(selectedData).toBeDefined();
      expect(Array.isArray(selectedData)).toBe(true);
      expect(selectedData.length).toBe(1);

      const record = selectedData[0];
      expect(record!.required_integer).toBe(42);
      expect(record!.required_varchar_maxlength).toBe("test value");
      expect(record!.required_char_fixed).toBe("abcde");
      expect(record!.required_enum).toBe("OPTION_A");
    });

    it("should insert with default values", async () => {
      const dataWithDefaults: ValidatorTestTable = {
        ...validData,
        optional_integer_with_default: undefined,
        optional_varchar_pattern: undefined,
        optional_char_with_default: undefined,
        optional_boolean_with_default: undefined,
        optional_uuid_with_default: undefined,
        optional_json_with_default: undefined,
        optional_numeric_with_default: undefined,
        optional_real_with_default: undefined,
        optional_double_precision_with_default: undefined,
        optional_timestamp_with_default: undefined,
        optional_date_with_default: undefined,
        optional_time_with_default: undefined,
        optional_enum_with_default: undefined,
        optional_point_with_default: undefined,
        optional_line_with_default: undefined,
        optional_interval_with_default: undefined,
      };

      const insertResult = await orm.validator_test_table
        .insert(dataWithDefaults)
        .returning(orm.validator_test_table.optional_integer_with_default, orm.validator_test_table.optional_varchar_pattern) as Array<Partial<ValidatorTestTable>>;

      expect(insertResult).toBeDefined();
      expect(insertResult[0]?.optional_integer_with_default).toBe(42);
    });

    it("should update existing records", async () => {
      // Insert first
      const insertResult = await orm.validator_test_table
        .insert(validData)
        .returning(orm.validator_test_table.id) as Array<Partial<ValidatorTestTable>>;

      const recordId = insertResult[0]?.id;
      expect(recordId).toBeDefined();

      // Update
      const updateResult = await orm.validator_test_table
        .update({
          required_integer: 999,
          required_varchar_maxlength: "updated value"
        })
        .where(new Equal(orm.validator_test_table.id, recordId))
        .returning(orm.validator_test_table.required_integer, orm.validator_test_table.required_varchar_maxlength) as Array<Partial<ValidatorTestTable>>;

      expect(updateResult).toBeDefined();
      expect(updateResult[0]?.required_integer).toBe(999);
      expect(updateResult[0]?.required_varchar_maxlength).toBe("updated value");
    });

    it("should delete records", async () => {
      // Insert first
      const insertResult = await orm.validator_test_table
        .insert(validData)
        .returning(orm.validator_test_table.id) as Array<Partial<ValidatorTestTable>>;

      const recordId = insertResult[0]?.id;
      expect(recordId).toBeDefined();

      // Delete
      const deleteResult = await orm.validator_test_table
        .delete()
        .where(new Equal(orm.validator_test_table.id, recordId))
        .returning(orm.validator_test_table.id) as Array<Partial<ValidatorTestTable>>;

      expect(deleteResult).toBeDefined();
      expect(deleteResult[0]?.id).toBe(recordId);

      // Verify deletion
      const verifyResult = await orm.validator_test_table
        .select()
        .where(new Equal(orm.validator_test_table.id, recordId));

      expect(verifyResult.length).toBe(0);
    });

    it("should handle multiple inserts", async () => {
      const multipleData = [
        { ...validData, required_integer: 100 },
        { ...validData, required_integer: 200, required_uuid: "660e8400-e29b-41d4-a716-446655440001" },
        { ...validData, required_integer: 300, required_uuid: "770e8400-e29b-41d4-a716-446655440002" }
      ];

      const insertResult = await orm.validator_test_table
        .insert(multipleData)
        .returning(orm.validator_test_table.id, orm.validator_test_table.required_integer) as Array<Partial<ValidatorTestTable>>;

      expect(insertResult).toBeDefined();
      expect(insertResult.length).toBe(3);
      expect(insertResult[0]?.required_integer).toBe(100);
      expect(insertResult[1]?.required_integer).toBe(200);
      expect(insertResult[2]?.required_integer).toBe(300);
    });

    it("should order and limit results", async () => {
      // Insert multiple records
      const multipleData = [
        { ...validData, required_integer: 100, required_varchar_maxlength: "zebra" },
        { ...validData, required_integer: 200, required_uuid: "660e8400-e29b-41d4-a716-446655440001", required_varchar_maxlength: "alpha" },
        { ...validData, required_integer: 300, required_uuid: "770e8400-e29b-41d4-a716-446655440002", required_varchar_maxlength: "beta" }
      ];

      await orm.validator_test_table.insert(multipleData);

      // Test ordering by integer descending
      const orderedByInt = await orm.validator_test_table
        .select(orm.validator_test_table.required_integer)
        .orderBy(orm.validator_test_table.required_integer, "DESC")
        .limit(2) as Array<Partial<ValidatorTestTable>>;

      expect(orderedByInt.length).toBe(2);
      expect(orderedByInt[0]?.required_integer).toBe(300);
      expect(orderedByInt[1]?.required_integer).toBe(200);

      // Test ordering by varchar ascending
      const orderedByVarchar = await orm.validator_test_table
        .select(orm.validator_test_table.required_varchar_maxlength)
        .orderBy(orm.validator_test_table.required_varchar_maxlength, "ASC")
        .limit(2) as Array<ValidatorTestTable>;

      expect(orderedByVarchar.length).toBe(2);
      expect(orderedByVarchar[0]?.required_varchar_maxlength).toBe("alpha");
      expect(orderedByVarchar[1]?.required_varchar_maxlength).toBe("beta");
    });

    it("should handle complex where conditions", async () => {
      // Insert test data
      const testData = [
        { ...validData, required_integer: 100, required_boolean: true },
        { ...validData, required_integer: 200, required_uuid: "660e8400-e29b-41d4-a716-446655440001", required_boolean: false },
        { ...validData, required_integer: 300, required_uuid: "770e8400-e29b-41d4-a716-446655440002", required_boolean: true }
      ];

      await orm.validator_test_table.insert(testData);

      // Test IN operator
      const inResult = await orm.validator_test_table
        .select(orm.validator_test_table.required_integer)
        .where(new In(orm.validator_test_table.required_integer, [100, 300]));

      expect(inResult.length).toBe(2);

      // Test AND condition (simulated by multiple queries for this test)
      const trueResults = await orm.validator_test_table
        .select(orm.validator_test_table.required_integer)
        .where(new Equal(orm.validator_test_table.required_boolean, true));

      expect(trueResults.length).toBe(2);
    });

    it("should handle special data types correctly", async () => {
      const specialData = {
        ...validData,
        required_uuid: "123e4567-e89b-12d3-a456-426614174000",
        required_json: { nested: { deeply: { value: "test" } }, array: [1, 2, 3] },
        required_timestamp: new Date("2024-01-15T10:30:00Z"),
        required_date: "2024-01-15",
        required_time: "23:45:30",
        required_point: new Point(-12.5, 8.75),
        required_line: new Line(2, -3, 5),
        required_interval: "1 year 2 months 3 days 4 hours 5 minutes 6 seconds"
      };

      const insertResult = await orm.validator_test_table
        .insert(specialData)
        .returning(
          orm.validator_test_table.required_uuid,
          orm.validator_test_table.required_json,
          orm.validator_test_table.required_timestamp,
          orm.validator_test_table.required_date,
          orm.validator_test_table.required_time,
          orm.validator_test_table.required_point,
          orm.validator_test_table.required_line,
          orm.validator_test_table.required_interval
        );

      expect(insertResult.length).toBe(1);
      const record = insertResult[0] as Partial<ValidatorTestTable>;

      expect(record.required_uuid).toBe("123e4567-e89b-12d3-a456-426614174000");
      expect(typeof record.required_json).toBe("object");
      expect(record.required_timestamp).toBeInstanceOf(Date);
      expect(record.required_time).toBe("23:45:30");

      // Note: Geometric types and intervals might be returned as strings from postgres
      expect(record.required_point).toBeDefined();
      expect(record.required_line).toBeDefined();
      expect(record.required_interval).toBeDefined();
    });
  });

  describe("Error Handling", () => {
    it("should handle constraint violations gracefully", async () => {
      const invalidData = {
        ...validData,
        required_uuid: "invalid-uuid-format"
      };

      expect(async () => {
        await orm.validator_test_table.insert(invalidData);
      }).toThrow();
    });
  });
});