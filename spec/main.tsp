import "@typespec/http";
import "@typespec/openapi";
import "@typespec/openapi3";

using Http;
using OpenAPI;

/**
 * API documentation for public facility reservation system.
 */
@service(#{ title: "Facility Reservation API" })
@info(#{ version: "1.0.0" })
namespace FacilityReservationAPI;

@error
model UnexpectedError {
  code: "INTERNAL_SERVER_ERROR";
  message: string;
}

model ProblemDetails {
  @header("content-type") contentType: "application/problem+json" | "application/json";

  /**
   * A URI reference [RFC3986] that identifies the problem type.
   */
  type?: string;

  /**
   * A short, human-readable summary of the problem type.
   */
  title?: string;

  /**
   * The HTTP status code for this occurrence of the problem.
   */
  status?: integer;

  /**
   * A human-readable explanation specific to this occurrence of the problem.
   */
  detail?: string;

  /**
   * A URI reference that identifies the specific occurrence of the problem.
   */
  instance?: string;
}

@format("email")
@maxLength(254)
scalar EmailString extends string;

/**
 * Serializer for admin-level access to user objects, including staff status and hyperlinked self-reference.
 */
model AdminUser {
  @visibility(Lifecycle.Read)
  id: integer;
  @visibility(Lifecycle.Read)
  url: url;

  @doc("Required. 150 characters or fewer. Letters, digits and @/./+/-/_ only.")
  @maxLength(150)
  @pattern("^[\\w.@+-]+$")
  username: string;

  @summary("Email address")
  email?: EmailString;

  /**
   * Designates whether the user can log into this admin site.
   */
  @summary("Staff status") is_staff?: boolean;
}

/**
 * Serializer for representing the currently authenticated user.
 */
model CurrentUser {
  @visibility(Lifecycle.Read)
  id: integer;

  @doc("Required. 150 characters or fewer. Letters, digits and @/./+/-/_ only.")
  @maxLength(150)
  @pattern("^[\\w.@+-]+$")
  username: string;

  @maxLength(254)
  @format("email")
  @summary("Email address")
  email?: string;
}

model PublicFacility {
  @visibility(Lifecycle.Read)
  id: integer;

  /**
   * Display name of the facility. Visible to users.
   */
  @maxLength(100) name: string;

  /**
   * Optional description of the facility, including usage rules or details.
   */
  description?: string;

  /**
   * Physical location or address of the facility.
   */
  @maxLength(255) location?: string;

  /**
   * Display priority. Lower numbers appear earlier in sorted lists.
   */
  @minValue(0) priority?: int64;

  /**
   * Set to false to disable this facility from public listing or reservation.
   */
  is_active?: boolean;

  @visibility(Lifecycle.Read)
  created_at?: utcDateTime;
  @visibility(Lifecycle.Read)
  updated_at?: utcDateTime;
}

/**
 * Retrieves a list of all registered users. Admin access required.
 */
@tag("admin")
@route("/api/v1/admin/users/")
@get
@summary("List all users")
op admin_users_list(): UnauthorizedResponse & ProblemDetails | ForbiddenResponse & ProblemDetails | Body<AdminUser[]> | UnexpectedError;

/**
 * Create a new user account. Admin access required.
 */
@tag("admin")
@route("/api/v1/admin/users/")
@post
@summary("Create a new user")
op admin_users_create(
  @header
  contentType: "application/json",
  @body body: AdminUser,
): BadRequestResponse & ProblemDetails | NotFoundResponse & ProblemDetails | CreatedResponse & AdminUser | UnexpectedError;

/**
 * Delete the user with the given ID. Admin access required.
 */
@tag("admin")
@route("/api/v1/admin/users/{id}/")
@delete
@summary("Delete a user (admin only)")
op admin_users_destroy(
  /**
   * A unique integer value identifying this user.
   */
  @path id: integer,
): UnauthorizedResponse & ProblemDetails | ForbiddenResponse & ProblemDetails | BadRequestResponse & ProblemDetails | NotFoundResponse & ProblemDetails | NoContentResponse | UnexpectedError;

/**
 * Fetch details of a specific user by ID. Admin access required.
 */
@tag("admin")
@route("/api/v1/admin/users/{id}/")
@get
@summary("Retrieve a user by ID")
op admin_users_retrieve(
  /**
   * A unique integer value identifying this user.
   */
  @path id: integer,
): UnauthorizedResponse & ProblemDetails | ForbiddenResponse & ProblemDetails | NotFoundResponse & ProblemDetails | AdminUser | UnexpectedError;

/**
 * Update select fields of a user. Admin access required.
 */
@tag("admin")
@route("/api/v1/admin/users/{id}/")
@patch
@summary("Partially update a user")
op admin_users_partial_update(
  /**
   * A unique integer value identifying this user.
   */
  @path id: integer,

  @header
  contentType: "application/merge-patch+json",
  @body body: MergePatchUpdate<AdminUser>,
): UnauthorizedResponse & ProblemDetails | ForbiddenResponse & ProblemDetails | BadRequestResponse & ProblemDetails | NotFoundResponse & ProblemDetails | AdminUser | UnexpectedError;

/**
 * Update an existing user's full data by ID. Admin access required.
 */
@tag("admin")
@route("/api/v1/admin/users/{id}/")
@put
@summary("Update a user")
op admin_users_update(
  /**
   * A unique integer value identifying this user.
   */
  @path id: integer,

  @header
  contentType: "application/json",
  @body body: AdminUser,
): BadRequestResponse & ProblemDetails | NotFoundResponse & ProblemDetails | AdminUser | UnexpectedError;

/**
 * Returns a list of all active facilities. No authentication required.
 */
@tag("facilities")
@route("/api/v1/facilities/")
@get
@summary("List all public facilities")
op facilities_list(): Body<PublicFacility[]> | UnexpectedError;

/**
 * Creates a new facility. Only administrators are authorized.
 */
@tag("facilities")
@route("/api/v1/facilities/")
@post
@summary("Create a facility (admin only)")
op facilities_create(
  @header
  contentType: "application/json",
  @body body: PublicFacility,
): ForbiddenResponse & ProblemDetails | BadRequestResponse & ProblemDetails | CreatedResponse & PublicFacility | UnexpectedError;

/**
 * Deletes a facility. Only administrators are authorized.
 */
@tag("facilities")
@route("/api/v1/facilities/{id}/")
@delete
@summary("Delete a facility (admin only)")
op facilities_destroy(
  /**
   * A unique integer value identifying this Facility.
   */
  @path id: integer,
): BadRequestResponse & ProblemDetails | NotFoundResponse & ProblemDetails | NoContentResponse | UnexpectedError;

/**
 * Returns detailed information about a specific facility. No authentication required.
 */
@tag("facilities")
@route("/api/v1/facilities/{id}/")
@get
@summary("Retrieve facility details")
op facilities_retrieve(
  /**
   * A unique integer value identifying this Facility.
   */
  @path id: integer,
): NotFoundResponse & ProblemDetails | PublicFacility | UnexpectedError;

/**
 * Updates select fields of a facility. Only administrators are authorized.
 */
@tag("facilities")
@route("/api/v1/facilities/{id}/")
@patch
@summary("Partially update a facility (admin only)")
op facilities_partial_update(
  /**
   * A unique integer value identifying this Facility.
   */
  @path id: integer,

  @header
  contentType: "application/merge-patch+json" | "application/json",
  @body body: MergePatchUpdate<PublicFacility>,
): BadRequestResponse & ProblemDetails | NotFoundResponse & ProblemDetails | PublicFacility | UnexpectedError;

/**
 * Updates an existing facility. Only administrators are authorized.
 */
@tag("facilities")
@route("/api/v1/facilities/{id}/")
@put
@summary("Update a facility (admin only)")
op facilities_update(
  /**
   * A unique integer value identifying this Facility.
   */
  @path id: integer,

  @header
  contentType: "application/json",
  @body body: PublicFacility,
): BadRequestResponse & ProblemDetails | NotFoundResponse & ProblemDetails | PublicFacility | UnexpectedError;

/**
 * Returns basic profile information of the currently authenticated user.
 */
@tag("me")
@route("/api/v1/me/")
@get
@summary("Retrieve current authenticated user")
op me_retrieve(): UnauthorizedResponse & ProblemDetails | CurrentUser | UnexpectedError;
