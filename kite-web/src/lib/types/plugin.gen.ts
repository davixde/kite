// Code generated by tygo. DO NOT EDIT.

//////////
// source: config.go

export type ConfigValues = { [key: string]: Record<string, any> | null};
export interface Config {
  sections: ConfigSection[];
}
export interface ConfigSection {
  name: string;
  description: string;
  fields: ConfigField[];
}
export interface ConfigField {
  key: string;
  type: ConfigFieldType;
  item_type: ConfigFieldType;
  name: string;
  description: string;
  required: boolean;
}
export type ConfigFieldType = string;
export const ConfigFieldTypeString: ConfigFieldType = "string";
export const ConfigFieldTypeInt: ConfigFieldType = "int";
export const ConfigFieldTypeBool: ConfigFieldType = "bool";
export const ConfigFieldTypeFloat: ConfigFieldType = "float";
export const ConfigFieldTypeArray: ConfigFieldType = "array";

//////////
// source: plugin.go

export type Plugin = any;
export type PluginInstance = any;
export interface Metadata {
  name: string;
  description: string;
  icon: string;
  author: string;
}
export interface Event {
  id: string;
  source: EventSource;
  type: EventType;
  description: string;
}
export type EventSource = string;
export const EventSourceDiscord: EventSource = "discord";
export type EventType = string;
export const EventTypeMessageCreate: EventType = "message_create";
export const EventTypeMessageReactionAdd: EventType = "message_reaction_add";
export const EventTypeMessageReactionRemove: EventType = "message_reaction_remove";
export interface Command {
  id: string;
  data: any /* api.CreateCommandData */;
}
