# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: bridge.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


import common_pb2 as common__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='bridge.proto',
  package='pb',
  syntax='proto3',
  serialized_options=b'Z\007../grpc',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n\x0c\x62ridge.proto\x12\x02pb\x1a\x0c\x63ommon.proto\"\x1e\n\x0fSayHelloRequest\x12\x0b\n\x03msg\x18\x01 \x01(\t\"\x1f\n\x10SayHelloResponse\x12\x0b\n\x03msg\x18\x01 \x01(\t\"F\n\x10TrainStepRequest\x12\x16\n\x04\x42\x61se\x18\x01 \x01(\x0b\x32\x08.pb.Base\x12\x1a\n\x06\x41\x63tion\x18\x02 \x01(\x0b\x32\n.pb.Action\"e\n\x11TrainStepResponse\x12\x16\n\x04\x42\x61se\x18\x01 \x01(\x0b\x32\x08.pb.Base\x12\x18\n\x05State\x18\x02 \x01(\x0b\x32\t.pb.State\x12\x1e\n\x08\x46\x65\x65\x64\x62\x61\x63k\x18\x03 \x01(\x0b\x32\x0c.pb.Feedback\")\n\x0fResetEnvRequest\x12\x16\n\x04\x42\x61se\x18\x01 \x01(\x0b\x32\x08.pb.Base\"D\n\x10ResetEnvResponse\x12\x16\n\x04\x42\x61se\x18\x01 \x01(\x0b\x32\x08.pb.Base\x12\x18\n\x05State\x18\x02 \x01(\x0b\x32\t.pb.State2\xb8\x01\n\x08TrainApi\x12\x37\n\x08SayHello\x12\x13.pb.SayHelloRequest\x1a\x14.pb.SayHelloResponse\"\x00\x12:\n\tTrainStep\x12\x14.pb.TrainStepRequest\x1a\x15.pb.TrainStepResponse\"\x00\x12\x37\n\x08ResetEnv\x12\x13.pb.ResetEnvRequest\x1a\x14.pb.ResetEnvResponse\"\x00\x42\tZ\x07../grpcb\x06proto3'
  ,
  dependencies=[common__pb2.DESCRIPTOR,])




_SAYHELLOREQUEST = _descriptor.Descriptor(
  name='SayHelloRequest',
  full_name='pb.SayHelloRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='msg', full_name='pb.SayHelloRequest.msg', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=34,
  serialized_end=64,
)


_SAYHELLORESPONSE = _descriptor.Descriptor(
  name='SayHelloResponse',
  full_name='pb.SayHelloResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='msg', full_name='pb.SayHelloResponse.msg', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=66,
  serialized_end=97,
)


_TRAINSTEPREQUEST = _descriptor.Descriptor(
  name='TrainStepRequest',
  full_name='pb.TrainStepRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='Base', full_name='pb.TrainStepRequest.Base', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='Action', full_name='pb.TrainStepRequest.Action', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=99,
  serialized_end=169,
)


_TRAINSTEPRESPONSE = _descriptor.Descriptor(
  name='TrainStepResponse',
  full_name='pb.TrainStepResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='Base', full_name='pb.TrainStepResponse.Base', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='State', full_name='pb.TrainStepResponse.State', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='Feedback', full_name='pb.TrainStepResponse.Feedback', index=2,
      number=3, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=171,
  serialized_end=272,
)


_RESETENVREQUEST = _descriptor.Descriptor(
  name='ResetEnvRequest',
  full_name='pb.ResetEnvRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='Base', full_name='pb.ResetEnvRequest.Base', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=274,
  serialized_end=315,
)


_RESETENVRESPONSE = _descriptor.Descriptor(
  name='ResetEnvResponse',
  full_name='pb.ResetEnvResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='Base', full_name='pb.ResetEnvResponse.Base', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='State', full_name='pb.ResetEnvResponse.State', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=317,
  serialized_end=385,
)

_TRAINSTEPREQUEST.fields_by_name['Base'].message_type = common__pb2._BASE
_TRAINSTEPREQUEST.fields_by_name['Action'].message_type = common__pb2._ACTION
_TRAINSTEPRESPONSE.fields_by_name['Base'].message_type = common__pb2._BASE
_TRAINSTEPRESPONSE.fields_by_name['State'].message_type = common__pb2._STATE
_TRAINSTEPRESPONSE.fields_by_name['Feedback'].message_type = common__pb2._FEEDBACK
_RESETENVREQUEST.fields_by_name['Base'].message_type = common__pb2._BASE
_RESETENVRESPONSE.fields_by_name['Base'].message_type = common__pb2._BASE
_RESETENVRESPONSE.fields_by_name['State'].message_type = common__pb2._STATE
DESCRIPTOR.message_types_by_name['SayHelloRequest'] = _SAYHELLOREQUEST
DESCRIPTOR.message_types_by_name['SayHelloResponse'] = _SAYHELLORESPONSE
DESCRIPTOR.message_types_by_name['TrainStepRequest'] = _TRAINSTEPREQUEST
DESCRIPTOR.message_types_by_name['TrainStepResponse'] = _TRAINSTEPRESPONSE
DESCRIPTOR.message_types_by_name['ResetEnvRequest'] = _RESETENVREQUEST
DESCRIPTOR.message_types_by_name['ResetEnvResponse'] = _RESETENVRESPONSE
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

SayHelloRequest = _reflection.GeneratedProtocolMessageType('SayHelloRequest', (_message.Message,), {
  'DESCRIPTOR' : _SAYHELLOREQUEST,
  '__module__' : 'bridge_pb2'
  # @@protoc_insertion_point(class_scope:pb.SayHelloRequest)
  })
_sym_db.RegisterMessage(SayHelloRequest)

SayHelloResponse = _reflection.GeneratedProtocolMessageType('SayHelloResponse', (_message.Message,), {
  'DESCRIPTOR' : _SAYHELLORESPONSE,
  '__module__' : 'bridge_pb2'
  # @@protoc_insertion_point(class_scope:pb.SayHelloResponse)
  })
_sym_db.RegisterMessage(SayHelloResponse)

TrainStepRequest = _reflection.GeneratedProtocolMessageType('TrainStepRequest', (_message.Message,), {
  'DESCRIPTOR' : _TRAINSTEPREQUEST,
  '__module__' : 'bridge_pb2'
  # @@protoc_insertion_point(class_scope:pb.TrainStepRequest)
  })
_sym_db.RegisterMessage(TrainStepRequest)

TrainStepResponse = _reflection.GeneratedProtocolMessageType('TrainStepResponse', (_message.Message,), {
  'DESCRIPTOR' : _TRAINSTEPRESPONSE,
  '__module__' : 'bridge_pb2'
  # @@protoc_insertion_point(class_scope:pb.TrainStepResponse)
  })
_sym_db.RegisterMessage(TrainStepResponse)

ResetEnvRequest = _reflection.GeneratedProtocolMessageType('ResetEnvRequest', (_message.Message,), {
  'DESCRIPTOR' : _RESETENVREQUEST,
  '__module__' : 'bridge_pb2'
  # @@protoc_insertion_point(class_scope:pb.ResetEnvRequest)
  })
_sym_db.RegisterMessage(ResetEnvRequest)

ResetEnvResponse = _reflection.GeneratedProtocolMessageType('ResetEnvResponse', (_message.Message,), {
  'DESCRIPTOR' : _RESETENVRESPONSE,
  '__module__' : 'bridge_pb2'
  # @@protoc_insertion_point(class_scope:pb.ResetEnvResponse)
  })
_sym_db.RegisterMessage(ResetEnvResponse)


DESCRIPTOR._options = None

_TRAINAPI = _descriptor.ServiceDescriptor(
  name='TrainApi',
  full_name='pb.TrainApi',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  create_key=_descriptor._internal_create_key,
  serialized_start=388,
  serialized_end=572,
  methods=[
  _descriptor.MethodDescriptor(
    name='SayHello',
    full_name='pb.TrainApi.SayHello',
    index=0,
    containing_service=None,
    input_type=_SAYHELLOREQUEST,
    output_type=_SAYHELLORESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='TrainStep',
    full_name='pb.TrainApi.TrainStep',
    index=1,
    containing_service=None,
    input_type=_TRAINSTEPREQUEST,
    output_type=_TRAINSTEPRESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='ResetEnv',
    full_name='pb.TrainApi.ResetEnv',
    index=2,
    containing_service=None,
    input_type=_RESETENVREQUEST,
    output_type=_RESETENVRESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
])
_sym_db.RegisterServiceDescriptor(_TRAINAPI)

DESCRIPTOR.services_by_name['TrainApi'] = _TRAINAPI

# @@protoc_insertion_point(module_scope)
