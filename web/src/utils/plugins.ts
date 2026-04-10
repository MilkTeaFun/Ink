import type { PluginFieldSpec } from "@/types/plugins";

export function getPluginFieldDefaultValue(field: PluginFieldSpec) {
  if (field.defaultValue !== undefined) {
    return field.defaultValue;
  }

  switch (field.type) {
    case "checkbox":
      return false;
    default:
      return "";
  }
}

export function buildPluginFieldDefaults(fields: PluginFieldSpec[]) {
  return fields.reduce<Record<string, unknown>>((accumulator, field) => {
    accumulator[field.key] = getPluginFieldDefaultValue(field);
    return accumulator;
  }, {});
}
