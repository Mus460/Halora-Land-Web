import { describe, expect, it } from "vitest";
import { rabAreaValue } from "./export-rekap-pdf";

describe("rabAreaValue", () => {
  it("prefers explicit area option", () => {
    expect(rabAreaValue({ buildingArea: "250" }, "120 m2")).toBe("120 m2");
  });
  it("uses buildingArea with m² suffix and id-ID format", () => {
    expect(rabAreaValue({ buildingArea: "245.5" })).toBe("245,5 m²");
    expect(rabAreaValue({ buildingArea: 250 })).toBe("250 m²");
  });
  it("falls back to dash when missing", () => {
    expect(rabAreaValue(undefined)).toBe("-");
    expect(rabAreaValue({ buildingArea: null })).toBe("-");
    expect(rabAreaValue({ buildingArea: 0 })).toBe("-");
  });
});
